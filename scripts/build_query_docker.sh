docker build -t query -f Dockerfile.query .

REGION="us-east-1"
REPO_URL=$(aws sts get-caller-identity --output text --query Account).dkr.ecr.$REGION.amazonaws.com/flowsys-query
aws ecr get-login-password --region=$REGION |  docker login --username AWS --password-stdin $(aws sts get-caller-identity --output text --query Account).dkr.ecr.$REGION.amazonaws.com
docker tag query:latest $REPO_URL
docker push $REPO_URL
