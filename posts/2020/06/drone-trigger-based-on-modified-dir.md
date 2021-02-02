---
title: How to trigger build steps based on modified directory?
date: Fri Jun  26 14:11:19 +07 2020
description: How Drone could handle monorepo?
categories:
    - DevOps
tags:
    - drone
    - ci
    - monorepo
---
Using monorepo with multiple micro services, every single commit will trigger a full lint/test/build/publish for every service. What can I do to limit the scope?

To do that, we can use `git diff` to show changes between commits:

```shell script
  if [[ -n "${DRONE_PULL_REQUEST}" ]]; then
    most_recent_before="origin/${DRONE_TARGET_BRANCH}"
  elif [[ "${DRONE_BUILD_EVENT}" = "push" && ("${DRONE_COMMIT_BRANCH}" = "master" || "${DRONE_COMMIT_BRANCH}" = "release-"*) ]]; then
    if [[ "${DRONE_COMMIT_BEFORE}" = "$zero" ]]; then
      exit 0
    else
      most_recent_before="${DRONE_COMMIT_BEFORE}"
    fi
  fi
  modified_files=$(git --no-pager diff --name-only "${DRONE_COMMIT_SHA}".."${most_recent_before}");
```

Moreover, if `go.mod` is changed, all affected modules should be re-tested:

```shell script
  if echo "$modified_files" | grep -q "go.mod"; then
    modified_deps=$(git diff --unified=0 "${DRONE_COMMIT_SHA}".."${most_recent_before}" go.mod | tail +6 | grep '^[+-]' | awk '{ print $2 }' | uniq)
    while read -r module; do
      if go list -f '{{ join .Deps "\n" }}' "./${module}" | grep -qf <(echo "${modified_deps}"); then
        echo "$module"
      fi
    done < <(echo "$all_modules_in_cluster")
  fi
```

Run this script, we will have a list of modified services:

```jsonnet
local listModifiedModules(cluster) = {
    name: "list-modified-modules",
    image: golang,
    volumes: mounts,
	commands: [
	    "./scripts/show_modified_folders.sh " + cluster + " > " + cluster + "/modified_modules.txt"
	],
};
```

Then the lint/test/build step can be run on that list:

```jsonnet
local lint(cluster) = {
	name: "lint",
	image: golangci,
	volumes: mounts,
	environment: {
	    "GOPACKAGESPRINTGOLISTERRORS": 1,
	},
	commands: [
		"./scripts/linter.sh " + cluster + "/modified_modules.txt"
	],
	depends_on: [
		"list-modified-modules"
	],
};
```

The last step we need to do is publish only changed services by using [drone-convert-pathschanged](https://github.com/meltwater/drone-convert-pathschanged) conversion extension:

```jsonnet
local publishPush(cluster, module) = {
	name: "publish-" + module,
	image: docker,
	settings: {
	},
	depends_on: [
		"build"
	],
	when: {
		branch: [
			"master",
			"release-*"
		],
		event: [
		    "push",
		],
		paths: {
		    include: [
		        cluster + "/" + module + "/**/*.go"
		    ],
		},
	}
};
```