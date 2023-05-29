package_name="GophKeeper"
build_directory="bin/"

version=$(git describe --always --long --dirty)
date=$(TZ=UTC date)
commit=$(git log -1 --pretty=format:"%H")

platforms=("windows/386" "windows/amd64" "darwin/amd64" "darwin/arm64" "linux/amd64" "linux/arm64")

for platform in "${platforms[@]}"
do
	platform_split=(${platform//\// })
	GOOS=${platform_split[0]}
	GOARCH=${platform_split[1]}
	output_name=$package_name'-'$GOOS'-'$GOARCH
	if [ $GOOS = "windows" ]; then
		output_name+='.exe'
	fi

	env GOOS=$GOOS GOARCH=$GOARCH go build -o $build_directory$output_name -ldflags "\
-X \"main.buildVersion=${version}\"\
-X \"main.buildDate=${date}\"\
-X \"main.buildCommit=${commit}\"" .

	if [ $? -ne 0 ]; then
   		echo 'An error has occurred! Aborting the script execution...'
		exit 1
	fi
done

