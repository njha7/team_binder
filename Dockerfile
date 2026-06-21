# TODO probably should find a more general purpose base image
# and layer on top of that.
FROM golang:1.26.3-trixie

# Create privileged user
RUN apt update && apt install -y sudo && \
  useradd -m -s /bin/bash devuser && \
  echo "devuser ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers && \
  chown -R devuser:devuser /home/devuser

USER devuser

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
ENV HOME="/home/devuser"
# Install  NVM
ENV BASH_ENV="${HOME}/.bash_env"
RUN touch "${BASH_ENV}"
RUN echo '. "${BASH_ENV}"' >> ~/.bashrc

RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.4/install.sh | PROFILE="${BASH_ENV}" bash
RUN echo node > .nvmrc
RUN nvm install

# Install Bazel
RUN npm install -g @bazel/bazelisk
RUN go install github.com/bazelbuild/buildtools/buildifier@latest

# Some conveniences + python
RUN sudo apt install -y python3-pip
RUN sudo apt install -y python3.13-venv
RUN sudo pip install --break-system-packages --quiet --no-input pip-tools

# Misc
RUN sudo pip install --break-system-packages --quiet --no-input black
RUN sudo pip install --break-system-packages --quiet --no-input pre-commit

# Coding Agents
RUN npm install -g opencode-ai
