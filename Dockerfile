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
# Add; iptables 1.4.21 (for strict compatability with GCI ChromiumOS iptables version)
RUN apk --update add iptables
COPY gci_iptables_conf_agent /
# OCI RunC standard requires numeric user id's
USER 0
ENTRYPOINT ["/gci_iptables_conf_agent"]
