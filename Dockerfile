FROM gcr.io/jenkinsxio/builder-go-maven:2.0.1062-389
USER root
ENV SONARQUBE_CLI_RELEASE_VERSION "4.2.0.1873"
ENV SHELLCHECK_RELEASE_VERSION "stable"

RUN yum install -y gcc-c++ make bzip2 GConf2 gtk2 chromedriver chromium xorg-x11-server-Xvfb && \
    curl -f --silent --location https://rpm.nodesource.com/setup_11.x | bash -  && \
    yum install -y nodejs  && \
    yum clean all
RUN npm i -g watch-cli vsce typescript aws-cdk

RUN wget "https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-${SONARQUBE_CLI_RELEASE_VERSION}.zip" && \
unzip "sonar-scanner-cli-${SONARQUBE_CLI_RELEASE_VERSION}.zip" && \
mv sonar-scanner-4.2.0.1873 /opt/sonar && \
rm -f sonar-scanner-cli-4.2.0.1873.zip && \
wget -qO- "https://storage.googleapis.com/shellcheck/shellcheck-${SHELLCHECK_RELEASE_VERSION}.linux.x86_64.tar.xz" | tar -xJv && \
cp "shellcheck-${SHELLCHECK_RELEASE_VERSION}/shellcheck" /usr/bin/ && \
rm -rf "shellcheck-${SHELLCHECK_RELEASE_VERSION}"

COPY ./build/jx-app-sonar-scanner /jx-app-sonar-scanner
COPY ./exec-sonar-scanner.sh /usr/local/bin/exec-sonar-scanner.sh
COPY ./sqproperties/ /sqproperties/
RUN ["chmod", "+x", "/usr/local/bin/exec-sonar-scanner.sh"]

ENTRYPOINT ["/jx-app-sonar-scanner"]

