variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "project_name" {
  description = "Project name used for resource naming"
  type        = string
  default     = "task-manager"
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "asia-southeast2"
}

variable "zone" {
  description = "GCP zone"
  type        = string
  default     = "asia-southeast2-a"
}

variable "environment" {
  description = "Environment (development, staging, production)"
  type        = string
  default     = "production"
}

variable "machine_type" {
  description = "GKE node machine type"
  type        = string
  default     = "e2-standard-2"
}

variable "min_node_count" {
  description = "Minimum number of nodes in node pool"
  type        = number
  default     = 2
}

variable "max_node_count" {
  description = "Maximum number of nodes in node pool"
  type        = number
  default     = 10
}

variable "db_tier" {
  description = "Cloud SQL instance tier"
  type        = string
  default     = "db-custom-2-4096"
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}
