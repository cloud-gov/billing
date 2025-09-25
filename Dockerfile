# This dockerfile is used for testing in CI only.
# In Cloud.gov, the app runs in a buildpack instead. For local development, a Docker container is not needed.
ARG base_image

FROM ${base_image}
COPY . /app
WORKDIR /app
