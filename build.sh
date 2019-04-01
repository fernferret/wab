#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# Setup Colors
green="\e[38;5;82m"
blue="\e[38;5;45m"
pink="\e[38;5;98m"
gold="\e[38;5;226m"
red="\e[38;5;160m"
gray="\e[38;5;245m"
default="\e[0m"

api="no"
js="no"
swagger="no"
gotags=()
anyerr=""
cd "$DIR"

buildmode=""
if [ "$1" != "" ]; then
  buildmode="$1"
elif [ "$BUILD_MODE" != "" ]; then
  buildmode="$BUILD_MODE"
else
  buildmode="all"
fi

debug=""
debug_goflags=""
if [ "$3" == "debug" ]; then
  debug="-gcflags=\"all=-N -l\" "
  debug_goflags="-X main.debug=debug "
fi

suffix=""
tiny=""

if [ "$buildmode" == "js" ]; then
  js="yes"
  swagger="yes"
elif [ "$buildmode" == "api" ]; then
  api="yes"
  swagger="yes"
elif [ "$buildmode" == "tiny" ]; then
  tiny="tiny "
  api="yes"
elif [ "$buildmode" == "all" ]; then
  swagger="yes"
  api="yes"
  js="yes"
elif [ "$buildmode" == "none" ]; then
  swagger="no"
  api="no"
  js="no"
else
  echo "Invalid argument."
  echo "usage: ./build.sh [js|api|tiny|all|none]"
  echo " or"
  echo "usage: BUILD_MODE=[js|api|tiny|all|none] ./build.sh"
  exit 1
fi
START=$(date +%s);
# Kickin' it with go 1.11, I want to make sure the GOPATH doesn't interfere
unset GOPATH

if [ "$js" == "yes" ]; then
  export NODE_ENV=production
  cd "$DIR/webapp"
  npm install
  npm run build
  if [ "$?" != "0" ]; then
    echo "Exiting due to javascript build errors..."
    exit 1
  fi
fi

# Only generate if we're doing JS or swagger
if [ "$js" == "yes" ] || [ "$swagger" == "yes" ]; then
  cd "$DIR"
  echo "Re-generating static files via vfsgen"
  go generate
  if [ "$?" != "0" ]; then
    echo "Exiting due to go generate failure..."
    exit 1
  fi
fi

mkdir -p bin

mode="${PROJECT_BUILD_MODE:-dev}"
build="wab"
builddisp="${build}${suffix}"
gitver="$(git describe --always --dirty --tags)"
# This is a neat snippet (getting only the latest tag) but I'm not using it anymore in favor of $mode
# gitver_short="$(git describe --abbrev=0 --tags)"
# if [ "${gitver_short}" != "${gitver}" ]; then
#   gitver_short="${gitver_short}-dev"
# fi

if [ "$CI_JOB_ID" != "" ]; then
  GITLAB_BUILD_VARS=" -X main.buildId=${CI_JOB_ID}"
fi

if [ "$swagger" == "yes" ]; then
  gotags+=("swagger")
fi

tags=""
if [ "$js" == "yes" ]; then
  echo ""
  ldflags="${debug}-ldflags='${debug_goflags}-X main.versionFull=${gitver}-full -X main.version=${gitver} -X main.versionMode=${mode}${GITLAB_BUILD_VARS}'"
  if [ ! ${#gotags[@]} -eq 0 ]; then
     tags="-tags '${gotags[@]}' "
  fi
  echo -e "${gray}Building complete binary as ${green}${builddisp}${pink} (with packaged vue app)${default}"
  build_str="go build $ldflags ${tags}-v -o bin/${builddisp} ./${build}"
  echo -e "${blue}=> ${pink}$build_str${default}"
  eval "$build_str"
  res=$?
  if [ "$res" != "0" ]; then
    echo -e "${blue} ==> [ ${pink}${builddisp}${blue} ] ${red}FAILED: ${gold}$res${default}"
    anyerr="yes"
  else
    sha1sum "bin/${builddisp}" > "bin/${builddisp}.sha1"
    echo -e "${blue} ==> [ ${pink}${builddisp}${blue} ] ${green}SUCCESS!${default} (sha1: $(cat bin/${builddisp}.sha1))"
  fi
fi

tags=""
if [ "$api" == "yes" ]; then
  echo ""
  gotags+=("apionly")
  echo "${gotags[@]}"
  if [ ! ${#gotags[@]} -eq 0 ]; then
     tags="-tags '${gotags[@]}' "
  fi
  builddisp="${build}-api${suffix}"
  ldflags="${debug}-ldflags='${debug_goflags}-X main.versionFull=${gitver}-api -X main.version=${gitver} -X main.versionMode=${mode}${GITLAB_BUILD_VARS}'"
  echo -e "${gray}Building ${tiny}api only binary as ${green}${builddisp}${default}"
  build_str="go build $ldflags ${tags}-v -o bin/${builddisp} ./${build}"
  echo -e "${blue}=> ${pink}$build_str${default}"
  eval "$build_str"
  res=$?
  if [ "$res" != "0" ]; then
    echo -e "${blue} ==> [ ${pink}${builddisp}${blue} ] ${red}FAILED: ${gold}$res${default}"
    anyerr="yes"
  else
    sha1sum "bin/${builddisp}" > "bin/${builddisp}.sha1"
    echo -e "${blue} ==> [ ${pink}${builddisp}${blue} ] ${green}SUCCESS!${default} (sha1: $(cat bin/${builddisp}.sha1))"
  fi
fi
if [ "$anyerr" == "yes" ]; then
  exit 1
fi
END=$(date +%s);
DIFF=$(( $END - $START ))
echo -e "${blue}Build took ${pink}$DIFF ${blue}seconds${default}"

