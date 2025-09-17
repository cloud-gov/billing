# This dockerfile is used for testing in CI only. The image runs in a buildpack in Cloud.gov and is developed locally without docker.
ARG base_image

FROM ${base_image}
COPY . /app
WORKDIR /app
