#!/bin/bash

export CLUSTER_NAME=eks-east-2-internal
export CLUSTER_REGION=us-east-2
export SERVICE_ACCOUNT_NAME=aws-s3-full
export PROFILE_NAMESPACE=default

eksctl create iamserviceaccount --name ${SERVICE_ACCOUNT_NAME} \
  --namespace ${PROFILE_NAMESPACE} \
  --cluster ${CLUSTER_NAME} \
  --region ${CLUSTER_REGION} \
  --attach-policy-arn=arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly \
  --attach-policy-arn=arn:aws:iam::aws:policy/AmazonS3FullAccess \
  --override-existing-serviceaccounts \
  --approve


cat <<EOF > secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-secret
  namespace: ${PROFILE_NAMESPACE}
  annotations:
    serving.kserve.io/s3-endpoint: s3.amazonaws.com
    serving.kserve.io/s3-usehttps: "1"
    serving.kserve.io/s3-region: ${CLUSTER_REGION}
type: Opaque
EOF

kubectl apply -f secret.yaml
kubectl patch serviceaccount ${SERVICE_ACCOUNT_NAME} -n ${PROFILE_NAMESPACE} -p '{"secrets": [{"name": "aws-secret"}]}'
rm -rf secret.yaml
