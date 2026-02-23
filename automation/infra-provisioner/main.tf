# ============================================================================
# Terraform Configuration for GCP Infrastructure
# Provisions: GKE Cluster, Cloud SQL, VPC, and supporting resources
# ============================================================================

terraform {
  required_version = ">= 1.5.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }

  # Remote state storage in GCS
  backend "gcs" {
    bucket = "hamfa-terraform-state"
    prefix = "task-manager/production"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# ── VPC Network ─────────────────────────────────────────────────
resource "google_compute_network" "main" {
  name                    = "${var.project_name}-vpc"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "main" {
  name          = "${var.project_name}-subnet"
  ip_cidr_range = "10.0.0.0/20"
  region        = var.region
  network       = google_compute_network.main.id

  secondary_ip_range {
    range_name    = "pods"
    ip_cidr_range = "10.1.0.0/16"
  }

  secondary_ip_range {
    range_name    = "services"
    ip_cidr_range = "10.2.0.0/20"
  }
}

# ── GKE Cluster ─────────────────────────────────────────────────
resource "google_container_cluster" "primary" {
  name     = "${var.project_name}-gke"
  location = var.zone

  # Use separately managed node pool
  remove_default_node_pool = true
  initial_node_count       = 1

  network    = google_compute_network.main.name
  subnetwork = google_compute_subnetwork.main.name

  ip_allocation_policy {
    cluster_secondary_range_name  = "pods"
    services_secondary_range_name = "services"
  }

  # Enable network policy
  network_policy {
    enabled = true
  }

  # Enable Workload Identity
  workload_identity_config {
    workload_pool = "${var.project_id}.svc.id.goog"
  }

  # Logging and monitoring
  logging_service    = "logging.googleapis.com/kubernetes"
  monitoring_service = "monitoring.googleapis.com/kubernetes"
}

resource "google_container_node_pool" "primary" {
  name       = "${var.project_name}-node-pool"
  location   = var.zone
  cluster    = google_container_cluster.primary.name
  node_count = var.min_node_count

  autoscaling {
    min_node_count = var.min_node_count
    max_node_count = var.max_node_count
  }

  node_config {
    machine_type = var.machine_type
    disk_size_gb = 50
    disk_type    = "pd-ssd"

    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform",
    ]

    labels = {
      environment = var.environment
      project     = var.project_name
    }

    # Enable workload identity on nodes
    workload_metadata_config {
      mode = "GKE_METADATA"
    }
  }

  management {
    auto_repair  = true
    auto_upgrade = true
  }
}

# ── Cloud SQL (PostgreSQL) ──────────────────────────────────────
resource "google_sql_database_instance" "postgres" {
  name             = "${var.project_name}-postgres"
  database_version = "POSTGRES_16"
  region           = var.region

  settings {
    tier              = var.db_tier
    availability_type = "REGIONAL" # High availability

    backup_configuration {
      enabled                        = true
      point_in_time_recovery_enabled = true
      start_time                     = "03:00"
    }

    ip_configuration {
      ipv4_enabled    = false
      private_network = google_compute_network.main.id
    }

    disk_size         = 20
    disk_type         = "PD_SSD"
    disk_autoresize   = true
  }

  deletion_protection = true
}

resource "google_sql_database" "taskmanager" {
  name     = "taskmanager"
  instance = google_sql_database_instance.postgres.name
}

resource "google_sql_user" "app_user" {
  name     = "hamfa"
  instance = google_sql_database_instance.postgres.name
  password = var.db_password
}

# ── Cloud Memorystore (Redis) ──────────────────────────────────
resource "google_redis_instance" "cache" {
  name           = "${var.project_name}-redis"
  tier           = "BASIC"
  memory_size_gb = 1
  region         = var.region

  authorized_network = google_compute_network.main.id

  redis_version = "REDIS_7_0"

  labels = {
    environment = var.environment
    project     = var.project_name
  }
}

# ── Cloud Storage (Backups) ─────────────────────────────────────
resource "google_storage_bucket" "backups" {
  name          = "${var.project_id}-backups"
  location      = var.region
  force_destroy = false

  versioning {
    enabled = true
  }

  lifecycle_rule {
    condition {
      age = 90
    }
    action {
      type = "Delete"
    }
  }

  lifecycle_rule {
    condition {
      age = 30
    }
    action {
      type          = "SetStorageClass"
      storage_class = "NEARLINE"
    }
  }
}

# ── Artifact Registry ──────────────────────────────────────────
resource "google_artifact_registry_repository" "docker" {
  location      = var.region
  repository_id = var.project_name
  format        = "DOCKER"

  labels = {
    environment = var.environment
  }
}
