# Release

* Update Makefile variable `VERSION` to the appropiate release version. Allowed formats:
  * alpha: `VERSION ?= 0.12.1-alpha.1`
  * stable: `VERSION ?= 0.12.1`

## Alpha

* If it is an **alpha** release, execute the following target to create appropiate `alpha` bundle files:

```bash
make prepare-alpha-release
```

* Then you can manually execute operator, bundle and catalog build/push targets.

```bash
make bundle-publish
```

```bash
make catalog-add-bundle-to-alpha
```

```bash
make catalog-publish
```

## Stable

* If it is an **stable** release, execute the following target to create appropiate `alpha` and `stable` bundle files:

```bash
make prepare-stable-release
```

* Then open a [Pull Request](https://github.com/3scale-sre/marin3r/pulls), and a GitHub Action will automatically detect if it is new release or not, in order to create it by building/pushing new operator and bundle images, as well as creating a GitHub release draft. Furthermore, an additional PR will automatically be opened (via the same "release" GH action) to publish the new updated catalog image. Review and approve the PR. Once merged, the changes will trigger the "catalog" GH action to build and push the new catalog image.
