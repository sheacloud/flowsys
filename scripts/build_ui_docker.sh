docker build -t ui -f Dockerfile.ui .

REGION="us-east-1"
REPO_URL=$(aws sts get-caller-identity --output text --query Account).dkr.ecr.$REGION.amazonaws.com/flowsys-ui
aws ecr get-login-password --region=$REGION |  docker login --username AWS --password-stdin $(aws sts get-caller-identity --output text --query Account).dkr.ecr.$REGION.amazonaws.com
docker tag ui:latest $REPO_URL
docker push $REPO_URL
