# Releasing

In order to release a new version of the library, just create a tag and push it.
```
export GTAG=v0.1.4
git tag -a ${GTAG} -m "Initial release"
git push origin ${GTAG}
```

# Vendoring
We are using [carvel's vendir](https://carvel.dev/vendir/) to vendor 3rd party resources that are needed for the tests.
We store these components in the `./vendoring` directory. To include/update components, add/update them in `./vendir.yml`
and call `vendir sync` from the root of the repo.

NOTE: as the e2e framework [cannot handle comment properly](https://github.com/kubernetes-sigs/e2e-framework/issues/388),
we need to remove the comments from the vendored resources:
```shell
yq '... comments=""' -i vendoring/gateway-api/experimental-install.yaml
```

