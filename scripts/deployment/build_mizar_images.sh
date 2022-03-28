#!/bin/bash

set -e

tag=goose
mizardir=/root/go/src/github.com/mizar
mizarbuilddir=$mizardir/etc/docker
slavehosts="./slaves.in"
buildlog="/tmp/build_image.out"
declare -a components=("mizar" "dropletd" "endpointopr")
declare -a dockerfiles=("mizar.Dockerfile" "daemon.Dockerfile" "operator.Dockerfile")
imagenamefile="/tmp/imgfiles.in"

build_image(){
    echo ">>>> building images"
    
    rm -f $buildlog >> $buildlog 2>&1
    pushd $mizardir >> $buildlog 2>&1
    length=${#dockerfiles[@]}
        
    echo "compiling mizar"
    make clean >> $buildlog 2>&1
    make all >> $buildlog 2>&1

    for (( i=0; i<${length}; i++ ));
    do
        dockerfilepath="$mizarbuilddir/${dockerfiles[$i]}"
        
	echo building vmizarnet/${components[$i]} using $dockerfilepath 

	image="vmizarnet/${components[$i]}:$tag"
	docker rmi -f $image >> $buildlog 2>&1
	docker image build -t $image -f $dockerfilepath . >> $buildlog 2>&1
    done

    popd 2>&1
}

zip_image_fn(){
    component=$1
    image="vmizarnet/$component:$tag"
    file="./$component.tar"
    echo saving $image to $file 

    rm -f $file
    docker save -o $file $image 
    gzip -f ${file} 
}

zip_image(){
    echo ">>>> zipping images"
    length=${#dockerfiles[@]}
    for (( i=0; i<${length}; i++ ));
    do
        zip_image_fn ${components[$i]} &
    done
    wait
}

send_file_names(){
    echo ">>>> reloading images on slave $slave"
    rm -f $imagenamefile
    length=${#components[@]}
    for (( i=0; i<${length}; i++ ));
    do
        echo ${components[$i]} >> $imagenamefile 
    done

    while IFS= read -r slave
    do
	echo $slave
        scp $imagenamefile $slave:/tmp >> $buildlog $2>&1
    done < "$slavehosts"
}

# ref https://stackoverflow.com/questions/22107610/shell-script-run-function-from-script-over-ssh
slave_reload_fn(){
    slavereloadlog=/tmp/reload.out
    rm -f $slavereloadlog
    imagenamefile=/tmp/imgfiles.in
    tag=$1

    # clean up dangling docker images
    docker rmi $(docker images --filter "dangling=true" -q --no-trunc) >> $slavereloadlog 2>&1

    gzip -d *.gz >> $slavereloadlog 2>&1

    while IFS= read -r component 
    do
        echo "reloading vmizarnet/$component:$tag with ./$component.tar"
	docker rmi -f vmizarnet/$component:$tag >> $slavereloadlog 2>&1
	docker load -i /root/$component.tar >> $slavereloadlog 2>&1
    done < "$imagenamefile"
    
    mv -f *.tar /tmp >> $slavereloadlog 2>&1
    mv -f *.gz /tmp >> $slavereloadlog 2>&1
}

send_and_reload(){
    slave=$1
    echo ">>>> reloading images on slave $slave"
 
    ssh $slave "rm -f *.tar.gz"
    scp *.tar.gz $slave:~ >> $buildlog 2>&1
    ssh $slave "$(typeset -f slave_reload_fn); slave_reload_fn $tag"
}

distribute_and_reload_images(){
    while IFS= read -r slave
    do
	send_and_reload $slave &
    done < "$slavehosts"
    wait
    rm -f *.tar.gz >> $buildlog 2>&1
}

echo ">> verifying slave host file exist"
if [ ! -f "$slavehosts" ]; then
    echo "$slavehosts does not exist."
    echo "put IPs of slave hosts in a file called **slave.in** and try again."
    echo "exited on purpose."
    exit
fi
    
build_image

zip_image

send_file_names

distribute_and_reload_images
