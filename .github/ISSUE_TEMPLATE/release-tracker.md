---
name: Release tracker
about: Create an issue to track release progress

---

<!-- < < < < < < < < < < < < < < < < < < < < < < < < < < < < < < < < < ☺ 
v                            ✰  Thanks for opening an issue! ✰    
v    Before smashing the submit button please review the template.
v    Word of caution: poorly thought-out proposals may be rejected 
v                     without deliberation 
☺ > > > > > > > > > > > > > > > > > > > > > > > > > > > > > > > > >  -->

## QA

- [ ] Tutorial tests verification
- [ ] Test `serve` on suite of chains

### Backwards compatibility

<!-- List of tests that need be performed with previous
versions of cli to guarantee that no regression is introduced -->

### Other testing

## Migration

<!-- Link to migration document -->

## Checklist

<!-- Remove any items that are not applicable. -->

- [ ] Update Ignite CLI version (see [#3793](https://github.com/ignite/cli/pull/3793) for example):
  - [ ] Rename module version in go.mod to `/vXX` (where `XX` is the new version number).
  - [ ] Update plugins go plush, protos and re-generate them
  - [ ] Update documentation links (docs/docs)
  - [ ] Update GitHub actions, goreleaser and other CI/CD scripts

## Post-release checklist

- [ ] Update [`changelog.md`](https://github.com/ignite/cli/blob/main/changelog.md)
- [ ] Update [`readme.md](https://github.com/ignite/cli/blob/main/readme.md):
  - [ ] Version matrix.
- [ ] Update docs site:
  - [ ] Add new release tag to [`docs/versioned_docs`](https://github.com/ignite/cli/tree/main/docs/versioned_docs).
- [ ] After changes to docs site are deployed, check [docs.ignite.com/](https://docs.ignite.com/) is updated.

____
