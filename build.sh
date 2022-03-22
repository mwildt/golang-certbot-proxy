#!/bin/bash

run_push=false
run_docker_build=false

while test $# -gt 0
do
    case "$1" in
        docker)
            run_docker_build=true
            ;;
        --push) 
            run_push=true
            ;;
        *) echo "ignore argumen $1"
            ;;
    esac
    shift
done

GO_BUILD="CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ."

echo "run $GO_BUILD"
eval $GO_BUILD


if [ "$run_docker_build" = true ] ; 
then

    echo "******************************"
    echo "* R U N   D O C K E R        *"
    echo "******************************"

    VERSION=$(date '+%Y-%m-%d-%H-%M-%S')

    BASE_IMAGE_NAME="maltewildt/golang-certbot-proxy"
    TAG="$BASE_IMAGE_NAME:$VERSION"
    LATEST="$BASE_IMAGE_NAME:latest"

    BUILD_CMD="docker build -t $TAG ."
    TAG_CMD="docker tag $TAG $LATEST"

    echo "run $BUILD_CMD"
    eval $BUILD_CMD

    echo "run $TAG_CMD"
    eval $TAG_CMD

    if [ "$run_push" = true ] ;
    then
        echo "try to push images"
        docker push $TAG
        docker push $LATEST
    else
        echo "pls run the following to push images to docker registry"
        echo "docker push $TAG"
        echo "docker push $LATEST"
    fi

fi



