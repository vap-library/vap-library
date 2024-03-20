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

