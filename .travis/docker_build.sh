#!/bin/bash
# -*- firestarter: "shfmt -i 4 -ci -w %p" -*-

set -euxo pipefail

readonly build_part=${BUILD_PART:-}
readonly gometalinter_path=$(readlink -f ./gometalinter.json)

build() {
    local build_dir="$1"
    pushd "$build_dir"
    gometalinter --config="$gometalinter_path" ./...
    go test -v -mod=vendor ./...
    popd

    for service in "${@:2}"; do
        docker build -f "docker-files/Dockerfile.$service" -t "kybernetwork/kyber-stats-$service:$TRAVIS_COMMIT" .
    done

}

case "$build_part" in
    1)
        build reserverates reserve-rates-api reserve-rates-crawler
        build users users-api
        build gateway gateway
        ;;
    2)
        build tradelogs trade-logs-api trade-logs-crawler
        build priceanalytics price-analytics-api
        ;;
    *)
        go test -v -mod=vendor $(go list -mod=vendor ./... | grep -v "github.com/KyberNetwork/reserve-stats/\(reserverates\|tradelogs\|users\|gateway\|priceanalytics\)")
        ;;
esac