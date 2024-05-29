# Releasing

In order to release a new version of the library, just create a tag and push it.
```
./release.sh
git commit -a -m 'message'
export GTAG=v0.1.4
git tag -a ${GTAG} -m "Initial release"
git push origin ${GTAG}
```

# Vendoring
We are using [carvel's vendir](https://carvel.dev/vendir/) to vendor 3rd party resources that are needed for the tests.
We store these components in the `./vendoring` directory. To include/update components, add/update them in `./vendir.yml`
and call `vendir sync` from the root of the repo.

NOTE: the stable release of the e2e framework [cannot handle comment on top of the yaml files properly](https://github.com/kubernetes-sigs/e2e-framework/issues/388)
Due to this we moved to a development release that includes the fix. Should there be a reason to switch back to the stable
version we need to remove the comments from the vendored resources:
```shell
yq '... comments=""' -i vendoring/gateway-api/experimental-install.yaml
```

# rootless podman issues with testing
Verify that you have all cgroups controllers
```
$ podman info --format json|jq '.host.cgroupControllers'
[
  "cpuset",
  "cpu",
  "io",
  "memory",
  "pids"
]
```

If not then make sure that you have the delegate configured:
```
$ cat /etc/systemd/system/user@.service.d/delegate.conf
[Service]
Delegate=yes
$ sudo systemctl daemon-reload
```

If still not then create a bash with systemd-run and run the tests from this bash:
```
$ systemd-run --scope --user bash
Running scope as unit: run-rd115f5ba51634ba898cf07275d493fe5.scope
```
