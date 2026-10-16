ARG BASE_IMAGE=ghcr.io/m1k1o/neko/base:latest
FROM $BASE_IMAGE

RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends ca-certificates curl gpg openbox; \
    #
    # install helium from its official signed apt repository (amd64/arm64)
    ARCH=$(dpkg --print-architecture); \
    curl -fsSL https://raw.githubusercontent.com/imputnet/helium-linux/main/pubkey.asc \
        | gpg --dearmor -o /usr/share/keyrings/helium.gpg; \
    echo "deb [arch=${ARCH} signed-by=/usr/share/keyrings/helium.gpg] https://pkg.helium.computer/deb stable main" \
        > /etc/apt/sources.list.d/helium.list; \
    apt-get update; \
    #
    # helium-bin only recommends packages, so install the libraries the
    # Chromium binary links against ourselves (names valid on Debian trixie)
    apt-get install -y --no-install-recommends helium-bin fonts-liberation libvulkan1 xdg-utils \
        libasound2t64 libatk-bridge2.0-0t64 libatk1.0-0t64 libatomic1 libatspi2.0-0t64 libcups2t64 \
        libdrm2 libgbm1 libgtk-3-0t64 libnspr4 libnss3 libpango-1.0-0 libxcomposite1 \
        libxdamage1 libxkbcommon0 libxrandr2 libxss1; \
    #
    # fail the build instead of the browser when a library is missing
    if ldd /opt/helium/helium | grep 'not found'; then exit 1; fi; \
    #
    # helium reads Chromium's policy directory; keep one helium-named path for
    # the policy mount in docker-compose.yaml
    mkdir -p /etc/helium/policies/managed /etc/chromium; \
    ln -s /etc/helium/policies /etc/chromium/policies; \
    #
    # clean up
    apt-get --purge autoremove -y gpg; \
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# copy configuation files
COPY supervisord.conf /etc/neko/supervisord/helium.conf
COPY policies.json /etc/helium/policies/managed/policies.json
COPY openbox.xml /etc/neko/openbox.xml
