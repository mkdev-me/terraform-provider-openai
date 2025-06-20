FROM hashicorp/terraform:light

ARG version=latest
ENV TERRAFORM_VERSION=$version

# Install necessary dependencies
RUN apk add --update --no-cache curl unzip bash jq

# Create working directory
RUN mkdir -p /root/terraform-provider-openai
WORKDIR /root/terraform-provider-openai

# Copy source and examples
COPY . .

# Install provider
RUN go build -o terraform-provider-openai
RUN mkdir -p /root/.terraform.d/plugins/github.com/fjcorp/openai/0.1.0/linux_amd64/
RUN cp terraform-provider-openai /root/.terraform.d/plugins/github.com/fjcorp/openai/0.1.0/linux_amd64/

# Copy provider file for legacy Terraform versions
RUN mkdir -p /root/.terraform.d/plugins/linux_amd64/
RUN cp terraform-provider-openai /root/.terraform.d/plugins/linux_amd64/terraform-provider-openai_v0.1.0

# Set up Terraform
WORKDIR /root/terraform-provider-openai/examples

# Use the local provider
COPY examples/provider.tf.docker /root/terraform-provider-openai/examples/provider.tf

# Initialize Terraform
RUN terraform init

# Default command
ENTRYPOINT ["terraform"]
CMD ["plan"] 