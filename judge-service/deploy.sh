#!/bin/bash
set -e

REGION="ap-northeast-1"
ACCOUNT_ID="127214181395"
REPO_NAME="judge-service-dev"
CLUSTER="j15-backend-cluster-dev"
SERVICE="judge-service-dev"

ECR_URI="${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${REPO_NAME}"

echo "=== Creating ECR repository ==="
aws ecr create-repository \
  --repository-name ${REPO_NAME} \
  --region ${REGION} 2>/dev/null || echo "Repository already exists"

echo "=== Logging into ECR ==="
aws ecr get-login-password --region ${REGION} | \
  docker login --username AWS --password-stdin ${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com

echo "=== Building Docker image ==="
docker build -t ${REPO_NAME} .

echo "=== Tagging and pushing image ==="
docker tag ${REPO_NAME}:latest ${ECR_URI}:latest
docker push ${ECR_URI}:latest

echo "=== Registering task definition ==="
aws ecs register-task-definition \
  --cli-input-json file://task-definition.json \
  --region ${REGION}

echo "=== Checking if service exists ==="
SERVICE_EXISTS=$(aws ecs describe-services \
  --cluster ${CLUSTER} \
  --services ${SERVICE} \
  --region ${REGION} \
  --query 'services[0].status' \
  --output text 2>/dev/null || echo "MISSING")

if [ "$SERVICE_EXISTS" = "ACTIVE" ]; then
  echo "=== Updating existing service ==="
  aws ecs update-service \
    --cluster ${CLUSTER} \
    --service ${SERVICE} \
    --task-definition ${SERVICE} \
    --force-new-deployment \
    --region ${REGION}
else
  echo "=== Creating new service ==="
  aws ecs create-service \
    --cluster ${CLUSTER} \
    --service-name ${SERVICE} \
    --task-definition ${SERVICE} \
    --desired-count 1 \
    --launch-type EC2 \
    --region ${REGION}
fi

echo "=== Deployment complete ==="
