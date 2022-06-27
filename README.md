## gh-discover

`gh extension install mszostok/gh-discover`

## Usage

### Discover user engagement

After installation, you can use the `gh discover engagement` command to collect all users, that were engaged in a given list of issue. By engagement, we mean:
- being an issuer author
- adding issue reaction
- adding comment under issue

#### Order by issues

```bash
GH_REPO=infracloudio/botkube \
gh discover engagement \
--issues=250,508,542 \
--ignore-users=mszostok \
--group-by=issues \
--cache-ttl 20m
```

>**Note**
> To print a raw Markdown format, add the `--raw` flag.

### Order by users

```bash
GH_REPO=infracloudio/botkube \
gh discover engagement \
--issues=250,508,542 \
--ignore-users=mszostok \
--group-by=users \
--cache-ttl 20m
```

>**Note**
> To print a raw Markdown format, add the `--raw` flag.
