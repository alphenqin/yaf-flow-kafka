# ================================
# 1. 构建阶段：编译 libfixbuf / YAF / super_mediator
# ================================
FROM debian:bookworm-slim AS builder

# 可选：构建时 HTTP 代理（给 curl / apt 用）
ARG APT_PROXY

ENV DEBIAN_FRONTEND=noninteractive
ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig

ARG FIXBUF_VERSION=3.0.0.alpha2
ARG YAF_VERSION=3.0.0.alpha4
ARG SM_VERSION=2.0.0.alpha3

# 配置阿里云源，并安装构建依赖（包含 libzmq3-dev，方便以后在镜像内编 Go）
RUN set -eux; \
    rm -f /etc/apt/sources.list.d/debian.sources || true; \
    echo "deb http://mirrors.aliyun.com/debian bookworm main contrib non-free non-free-firmware" > /etc/apt/sources.list; \
    echo "deb http://mirrors.aliyun.com/debian bookworm-updates main contrib non-free non-free-firmware" >> /etc/apt/sources.list; \
    echo "deb http://mirrors.aliyun.com/debian-security bookworm-security main contrib non-free non-free-firmware" >> /etc/apt/sources.list; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
        build-essential \
        pkg-config \
        wget curl ca-certificates \
        libglib2.0-dev \
        libpcap-dev \
        libpcre3-dev \
        zlib1g-dev \
        libssl-dev \
        liblua5.3-dev \
        libzmq3-dev \
    ; \
    rm -rf /var/lib/apt/lists/*

WORKDIR /tmp

# ----------------
# 1.1 编译 libfixbuf
# ----------------
RUN HTTP_PROXY=${APT_PROXY} HTTPS_PROXY=${APT_PROXY} \
    curl -fSL "https://tools.netsa.cert.org/releases/libfixbuf-${FIXBUF_VERSION}.tar.gz" -o libfixbuf.tar.gz && \
    tar xzf libfixbuf.tar.gz && \
    cd "libfixbuf-${FIXBUF_VERSION}" && \
    ./configure --disable-tools && \
    make -j"$(nproc)" && \
    make install && \
    ldconfig && \
    cd /tmp && \
    rm -rf "libfixbuf-${FIXBUF_VERSION}" libfixbuf.tar.gz

# ----------------
# 1.2 编译 YAF（开启 applabel + DPI）
# ----------------
RUN HTTP_PROXY=${APT_PROXY} HTTPS_PROXY=${APT_PROXY} \
    curl -fSL "https://tools.netsa.cert.org/releases/yaf-${YAF_VERSION}.tar.gz" -o yaf.tar.gz && \
    tar xzf yaf.tar.gz && \
    cd "yaf-${YAF_VERSION}" && \
    ./configure --enable-applabel --enable-dpi && \
    make -j"$(nproc)" && \
    make install && \
    ldconfig && \
    cd /tmp && \
    rm -rf "yaf-${YAF_VERSION}" yaf.tar.gz

# ----------------
# 1.3 编译 super_mediator
# ----------------
RUN HTTP_PROXY=${APT_PROXY} HTTPS_PROXY=${APT_PROXY} \
    curl -fSL "https://tools.netsa.cert.org/releases/super_mediator-${SM_VERSION}.tar.gz" -o sm.tar.gz && \
    tar xzf sm.tar.gz && \
    cd "super_mediator-${SM_VERSION}" && \
    ./configure --with-mysql=no && \
    make -j"$(nproc)" && \
    make install && \
    ldconfig && \
    cd /tmp && \
    rm -rf "super_mediator-${SM_VERSION}" sm.tar.gz

# ================================
# 2. 运行阶段：精简运行环境 + csv2kafka + 启动脚本
# ================================
FROM debian:bookworm-slim AS runtime

ENV DEBIAN_FRONTEND=noninteractive

# 配置阿里云源，安装运行时依赖（包含 libzmq5 + bash）
RUN set -eux; \
    rm -f /etc/apt/sources.list.d/debian.sources || true; \
    echo "deb http://mirrors.aliyun.com/debian bookworm main contrib non-free non-free-firmware" > /etc/apt/sources.list; \
    echo "deb http://mirrors.aliyun.com/debian bookworm-updates main contrib non-free non-free-firmware" >> /etc/apt/sources.list; \
    echo "deb http://mirrors.aliyun.com/debian-security bookworm-security main contrib non-free non-free-firmware" >> /etc/apt/sources.list; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        libglib2.0-0 \
        libpcap0.8 \
        libpcre3 \
        zlib1g \
        libssl3 \
        liblua5.3-0 \
        libzmq5 \
        bash \
    ; \
    rm -rf /var/lib/apt/lists/*

# 拷贝 builder 安装好的 libfixbuf + yaf + super_mediator
COPY --from=builder /usr/local/ /usr/local/

# 让动态链接器能找到 /usr/local/lib 的 so，并刷新缓存
RUN set -eux; \
    echo "/usr/local/lib"  >  /etc/ld.so.conf.d/usr-local.conf; \
    echo "/usr/local/lib64" >> /etc/ld.so.conf.d/usr-local.conf; \
    ldconfig

# （可选兜底）再显式加 LD_LIBRARY_PATH
ENV LD_LIBRARY_PATH=/usr/local/lib:/usr/local/lib64

# 拷贝你的 Go 二进制（需要你在构建上下文里准备好）
COPY csv2kafka /usr/local/bin/csv2kafka
RUN chmod +x /usr/local/bin/csv2kafka

# 准备 YAF 配置目录
RUN mkdir -p /etc/yaf

# 写入口脚本：使用官方 --config yaf.init 启动 YAF
RUN cat >/entrypoint.sh <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

# =========================
# 环境变量（docker run -e 覆盖）
# =========================

# YAF 官方配置文件路径（必须存在）
YAF_CONFIG_FILE="${YAF_CONFIG_FILE:-/etc/yaf/yaf.init}"

# super_mediator 相关
SM_LISTEN_PORT="${SM_LISTEN_PORT:-18000}"   # super_mediator 监听 IPFIX 端口，要和 yaf.init 里的 output.port 对应
SM_FIELDS="${SM_FIELDS:-flowStartMilliseconds,flowEndMilliseconds,sourceIPv4Address,destinationIPv4Address,sourceTransportPort,destinationTransportPort,protocolIdentifier,silkAppLabel}"

# csv2kafka 相关
CSV_TRANSPORT="${CSV_TRANSPORT:-kafka}"     # kafka / zmq
CSV_TOPIC="${CSV_TOPIC:-yaf_csv}"           # transport=kafka 时用
CSV_ZMQ_ENDPOINT="${CSV_ZMQ_ENDPOINT:-tcp://127.0.0.1:5555}"  # transport=zmq 时用

echo "[entrypoint] 使用 YAF 配置文件: ${YAF_CONFIG_FILE}"
echo "[entrypoint] super_mediator 监听端口: ${SM_LISTEN_PORT}"
echo "[entrypoint] TEXT 输出字段: ${SM_FIELDS}"
echo "[entrypoint] csv2kafka 传输方式: ${CSV_TRANSPORT}"

# 检查 YAF 配置文件是否存在
if [ ! -f "${YAF_CONFIG_FILE}" ]; then
  echo "[entrypoint] ERROR: 未找到 YAF 配置文件: ${YAF_CONFIG_FILE}" >&2
  echo "[entrypoint] 请在 docker run 时用 -v 挂载，比如：" >&2
  echo "  -v /path/on/host/yaf.init:/etc/yaf/yaf.init:ro" >&2
  exit 1
fi

# =========================
# 组装 csv2kafka 参数
# =========================
CSV_ARGS=(-transport "${CSV_TRANSPORT}" -skip-header=true)
if [ "${CSV_TRANSPORT}" = "kafka" ]; then
  CSV_ARGS+=(-topic "${CSV_TOPIC}")
elif [ "${CSV_TRANSPORT}" = "zmq" ]; then
  CSV_ARGS+=(-zmq-endpoint "${CSV_ZMQ_ENDPOINT}")
fi
echo "[entrypoint] csv2kafka 参数: ${CSV_ARGS[*]}"

# =========================
# super_mediator：作为 IPFIX collector，从 YAF 接收 → TEXT → csv2kafka(stdin)
# =========================
super_mediator \
  --ipfix-input=tcp \
  --ipfix-port="${SM_LISTEN_PORT}" \
  --output-mode=TEXT \
  --print-headers \
  --out=- \
  --fields="${SM_FIELDS}" \
  localhost \
  | /usr/local/bin/csv2kafka "${CSV_ARGS[@]}" &
SM_PIPE_PID=$!

echo "[entrypoint] super_mediator + csv2kafka pipeline 已启动 (PID=${SM_PIPE_PID})"

# =========================
# 启动 YAF（官方 --config yaf.init）
# =========================
yaf --config "${YAF_CONFIG_FILE}" "$@" &
YAF_PID=$!

echo "[entrypoint] YAF 已启动 (PID=${YAF_PID})，配置文件: ${YAF_CONFIG_FILE}"

# =========================
# 等 super_mediator + csv2kafka 退出，容器随之退出
# =========================
wait "${SM_PIPE_PID}"
EOF

RUN chmod +x /entrypoint.sh
RUN sed -i 's/\r$//' /entrypoint.sh || true

# 非 root 用户（抓网卡时 docker run 加 cap）
RUN useradd -r -s /usr/sbin/nologin yaf
USER yaf

ENTRYPOINT ["/entrypoint.sh"]
CMD []
