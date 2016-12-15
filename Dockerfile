#
# Copyright © 2016 Samsung CNCT
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License. 
#
# Dockerfile - GCI IP Tables Config Manager
#
# - Example Commands:
# docker build --rm -t sostheim/kube-gci-conf .
# docker run --rm -it --net=host --privileged sostheim/kube-gci-conf
#
# Alpine Linux 3.3 is required for correct apk versions
FROM alpine:3.3
MAINTAINER Rick Sostheim
LABEL vendor="Samsung CNCT"

# Add iptables 1.4.21 (current version in alpine:3.3)
RUN apk --update add iptables

# GCI is 1.4.21, Container VM is 1.4.14
# Indicate any irrelevant value as "*" => "wildcard / don't care"
ENV IPTABLES_MAJOR=1 IPTABLES_MINOR=4 IPTABLES_PATCH=* IPTABLES_VERSION=1.4.14

# Sleep interval - in seconds.
ENV IPTABLES_CHECK_INTERVAL=60

# OCI RunC standard requires numeric user id's
USER 0

COPY gci_iptables_conf_agent /
ENTRYPOINT ["/gci_iptables_conf_agent"]
