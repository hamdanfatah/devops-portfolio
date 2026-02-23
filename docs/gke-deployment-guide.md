# GKE Autopilot Deployment Guide ($5 Budget)

> **Tujuan:** Deploy Task Manager API ke GKE Autopilot dengan auto-scaling dan GitHub Actions CI/CD.
> **Estimasi biaya:** ~$1.50-2.00 untuk sesi 4 jam.
> **⚠️ PENTING:** Hapus semua resource setelah selesai agar credit tidak habis!

---

## Prerequisites

- GCP account dengan credit tersisa
- `gcloud` CLI terinstall ([install guide](https://cloud.google.com/sdk/docs/install))
- `kubectl` terinstall
- `helm` terinstall
- Docker terinstall (sudah ✅)

---

## Fase 1: Setup GCP Resources (~15 menit)

### 1.1 Login & Set Project

```bash
gcloud auth login
gcloud config set project global-exchange-487402-r4
```

### 1.2 Enable APIs

```bash
gcloud services enable \
  container.googleapis.com \
  artifactregistry.googleapis.com \
  cloudbuild.googleapis.com
```

### 1.3 Buat Artifact Registry (Tempat Simpan Docker Image)

```bash
gcloud artifacts repositories create task-manager \
  --repository-format=docker \
  --location=asia-southeast2 \
  --description="Task Manager Docker images"
```

### 1.4 Buat GKE Autopilot Cluster

```bash
gcloud container clusters create-auto task-manager-cluster \
  --region=asia-southeast2 \
  --project=global-exchange-487402-r4
```

⏱️ Ini butuh ~5-10 menit. Sabar!

### 1.5 Connect kubectl ke Cluster

```bash
gcloud container clusters get-credentials task-manager-cluster \
  --region=asia-southeast2

# Verifikasi:
kubectl get nodes
```

---

## Fase 2: Build & Push Docker Image (~10 menit)

### 2.1 Authenticate Docker ke Artifact Registry

```bash
gcloud auth configure-docker asia-southeast2-docker.pkg.dev
```

### 2.2 Build & Push Image

```bash
cd ~/devops-portfolio/app

# Build image
docker build -t asia-southeast2-docker.pkg.dev/global-exchange-487402-r4/task-manager/task-manager-api:v1 .

# Push ke registry
docker push asia-southeast2-docker.pkg.dev/global-exchange-487402-r4/task-manager/task-manager-api:v1
```

---

## Fase 3: Deploy ke GKE via Helm (~10 menit)

### 3.1 Update values.yaml untuk Autopilot

Buat file `values-autopilot.yaml` untuk override:

```bash
cat > ~/devops-portfolio/k8s/helm-chart/values-autopilot.yaml << 'EOF'
replicaCount: 1

image:
  repository: asia-southeast2-docker.pkg.dev/global-exchange-487402-r4/task-manager/task-manager-api
  tag: v1
  pullPolicy: Always

service:
  type: LoadBalancer
  port: 80
  targetPort: 8080

ingress:
  enabled: false

resources:
  requests:
    cpu: 250m
    memory: 512Mi
  limits:
    cpu: 500m
    memory: 512Mi

autoscaling:
  enabled: true
  minReplicas: 1
  maxReplicas: 5
  targetCPUUtilizationPercentage: 50

config:
  SERVER_PORT: "8080"
  ENVIRONMENT: "production"
  POSTGRES_HOST: "postgres-service"
  POSTGRES_PORT: "5432"
  POSTGRES_DB: "taskmanager"
  POSTGRES_SSL_MODE: "disable"
  MONGO_URI: "mongodb://mongodb-service:27017"
  MONGO_DB: "taskmanager"
  REDIS_HOST: "redis-service"
  REDIS_PORT: "6379"

secrets:
  POSTGRES_USER: "hamfa"
  POSTGRES_PASSWORD: "hamfa_secret"
EOF
```

### 3.2 Deploy Database Services Dulu

Untuk menghemat biaya, kita deploy database sebagai pod sederhana:

```bash
# PostgreSQL
kubectl run postgres --image=postgres:16-alpine \
  --env="POSTGRES_USER=hamfa" \
  --env="POSTGRES_PASSWORD=hamfa_secret" \
  --env="POSTGRES_DB=taskmanager" \
  --port=5432
kubectl expose pod postgres --name=postgres-service --port=5432

# MongoDB
kubectl run mongodb --image=mongo:4.4 --port=27017
kubectl expose pod mongodb --name=mongodb-service --port=27017

# Redis
kubectl run redis --image=redis:7-alpine --port=6379
kubectl expose pod redis --name=redis-service --port=6379

# Tunggu sampai semua running:
kubectl get pods -w
```

### 3.3 Deploy App via Helm

```bash
cd ~/devops-portfolio
helm install task-manager ./k8s/helm-chart/ \
  -f ./k8s/helm-chart/values-autopilot.yaml
```

### 3.4 Verifikasi Deployment

```bash
# Cek pods
kubectl get pods

# Cek services (tunggu EXTERNAL-IP muncul)
kubectl get svc

# Cek logs
kubectl logs -l app.kubernetes.io/name=task-manager

# Test API (ganti EXTERNAL_IP)
curl http://EXTERNAL_IP/health
```

---

## Fase 4: Test Auto-Scaling (~15 menit)

### 4.1 Cek HPA (Horizontal Pod Autoscaler)

```bash
kubectl get hpa
```

### 4.2 Simulasi Load (Trigger Auto-Scale)

Buka terminal baru dan jalankan:

```bash
# Kirim banyak request untuk trigger auto-scaling
while true; do curl -s http://EXTERNAL_IP/health > /dev/null; done
```

### 4.3 Monitor Scaling

Di terminal lain:

```bash
# Watch pods bertambah secara real-time
kubectl get pods -w

# Watch HPA metrics
kubectl get hpa -w
```

### 4.4 Test Self-Healing

```bash
# Hapus pod secara paksa
kubectl delete pod $(kubectl get pods -l app.kubernetes.io/name=task-manager -o jsonpath="{.items[0].metadata.name}")

# Lihat pod baru otomatis dibuat!
kubectl get pods -w
```

---

## Fase 5: Connect GitHub Actions (~20 menit)

### 5.1 Buat Service Account

```bash
# Buat service account
gcloud iam service-accounts create github-deploy \
  --display-name="GitHub Actions Deploy"

# Berikan izin
PROJECT_ID=$(gcloud config get-value project)

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:github-deploy@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/container.developer"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:github-deploy@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"
```

### 5.2 Setup Workload Identity Federation

```bash
# Buat pool
gcloud iam workload-identity-pools create github-pool \
  --location="global" \
  --display-name="GitHub Pool"

# Buat provider
gcloud iam workload-identity-pools providers create-oidc github-provider \
  --location="global" \
  --workload-identity-pool="github-pool" \
  --display-name="GitHub Provider" \
  --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository" \
  --issuer-uri="https://token.actions.githubusercontent.com"

# Bind service account
gcloud iam service-accounts add-iam-policy-binding \
  github-deploy@${PROJECT_ID}.iam.gserviceaccount.com \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/projects/$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')/locations/global/workloadIdentityPools/github-pool/attribute.repository/hamdanfatah/devops-portfolio"
```

### 5.3 Dapatkan WIF Provider String

```bash
gcloud iam workload-identity-pools providers describe github-provider \
  --location="global" \
  --workload-identity-pool="github-pool" \
  --format="value(name)"
```

Catat output-nya (format: `projects/NOMOR/locations/global/...`)

### 5.4 Masukkan GitHub Secrets

Buka: https://github.com/hamdanfatah/devops-portfolio/settings/secrets/actions

Tambahkan 3 secrets:
| Secret Name | Value |
|:---|:---|
| `GCP_PROJECT_ID` | Project ID kamu |
| `WIF_PROVIDER` | Output dari langkah 5.3 |
| `WIF_SERVICE_ACCOUNT` | `github-deploy@PROJECT_ID.iam.gserviceaccount.com` |

### 5.5 Test: Push Kode → Auto Deploy

```bash
# Ubah sesuatu kecil, commit, push
git add .
git commit -m "test: trigger CD pipeline"
git push
```

Buka GitHub → Actions → lihat CD pipeline jalan dan deploy ke GKE!

---

## Fase 6: Screenshot & Dokumentasi (~10 menit)

Ambil screenshot untuk portfolio:

1. **GKE Dashboard** — Console → Kubernetes Engine → Workloads
2. **Pods Running** — `kubectl get pods` menunjukkan pods healthy
3. **HPA Scaling** — `kubectl get hpa` menunjukkan auto-scaling aktif
4. **GitHub Actions** — Tab Actions menunjukkan CD berhasil ✅
5. **API Response** — Browser menampilkan health check dari EXTERNAL_IP

---

## 🚨 Fase 7: HAPUS SEMUA RESOURCES (WAJIB!)

```bash
# Hapus Helm release
helm uninstall task-manager

# Hapus database pods
kubectl delete pod postgres mongodb redis
kubectl delete svc postgres-service mongodb-service redis-service

# HAPUS CLUSTER (ini yang paling penting!)
gcloud container clusters delete task-manager-cluster \
  --region=asia-southeast2 --quiet

# Hapus Artifact Registry (opsional, biaya kecil)
gcloud artifacts repositories delete task-manager \
  --location=asia-southeast2 --quiet

# Hapus service account
gcloud iam service-accounts delete \
  github-deploy@${PROJECT_ID}.iam.gserviceaccount.com --quiet
```

> **⚠️ PASTIKAN cluster sudah terhapus!** Cek di Console → Kubernetes Engine → Clusters. Harus kosong!

---

## Estimasi Biaya Final

| Resource                      | Estimasi      |
| :---------------------------- | :------------ |
| GKE Autopilot pods (4 jam)    | ~$1.50        |
| Artifact Registry             | ~$0.10        |
| Network (LoadBalancer, 4 jam) | ~$0.30        |
| **Total**                     | **~$1.90**    |
| **Sisa credit**               | **~$3.10** ✅ |
