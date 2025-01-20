FROM ubuntu:24.04
ARG VERSION="dev"
ENV NO_PROXY_CACHE="-o Acquire::BrokenProxy=true -o Acquire::http::No-Cache=true -o Acquire::http::Pipeline-Depth=0"

ENV TMP_DIR="/tron-build"
ENV JDK_TAR="jdk-8u202-linux-x64.tar.gz"
ENV JDK_DIR="jdk1.8.0_202"
ENV JDK_MD5="0029351f7a946f6c05b582100c7d45b7"
ENV BASE_DIR="/java-tron"

# Update and install dependencies without using any cache
RUN apt-get update $NO_PROXY_CACHE  && \
  apt-get --quiet --yes install git wget 7zip curl jq && \
  wget -P /usr/local https://github.com/frekele/oracle-java/releases/download/8u202-b08/$JDK_TAR \
  && echo "$JDK_MD5 /usr/local/$JDK_TAR" | md5sum -c \
  && tar -zxf /usr/local/$JDK_TAR -C /usr/local\
  && rm /usr/local/$JDK_TAR \
  && export JAVA_HOME=/usr/local/$JDK_DIR \
  && export CLASSPATH=$JAVA_HOME/lib/dt.jar:$JAVA_HOME/lib/tools.jar \
  && export PATH=$PATH:$JAVA_HOME/bin \
  && echo "git clone" \
  && mkdir -p $TMP_DIR \
  && cd $TMP_DIR \
  && git clone https://github.com/tronprotocol/java-tron.git \
  && cd java-tron \
  && git checkout master \
  && ./gradlew build -x test \
  && cd build/distributions \
  && 7z x -y java-tron-1.0.0.zip  \
  && mv java-tron-1.0.0 $BASE_DIR \
  && rm -rf $TMP_DIR \
  && rm -rf ~/.gradle \
  && mv /usr/local/$JDK_DIR/jre /usr/local \
  && rm -rf /usr/local/$JDK_DIR \
  # Clean apt cache
  apt-get clean  && \
  rm -rf /var/cache/apt/archives/* /var/cache/apt/archives/partial/*  && \
  rm -rf /var/lib/apt/lists/*

RUN wget -P $BASE_DIR/config https://raw.githubusercontent.com/tronprotocol/tron-deployment/master/main_net_config.conf

ENV JAVA_HOME="/usr/local/jre"
ENV PATH=$PATH:$JAVA_HOME/bin

COPY docker-entrypoint.sh $BASE_DIR/bin

WORKDIR $BASE_DIR

ENTRYPOINT ["./bin/docker-entrypoint.sh"]

# Build-time metadata as defined at http://label-schema.org
ARG BUILD_DATE
ARG VCS_REF
LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.name="Java-TRON" \
      org.label-schema.description="TRON protocol" \
      org.label-schema.url="https://tron.network/" \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url="https://github.com/tronprotocol/java-tron.git" \
      org.label-schema.vendor="TRON protocol" \
      org.label-schema.version=$VERSION \
      org.label-schema.schema-version="1.0"