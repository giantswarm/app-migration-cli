# app-migration-cli

This tool migrates `apps` from one MC to another. It's closly coupled
to the vintage -> capi [migration tool](https://github.com/giantswarm/capi-migration-cli).

> :warning: WIP

## :airplane: Rundown

Parts of the script must run in advance of the CAPI Migration, others
are scheduled after a successfull infrastructure migration.

1. **preflight** - *readonly checks if a migration is possible*
    * validate access to both mcs
    * check WC condition/health (only "Created" is allowed)
    * :bulp: check `app` `extraConfig` config for a working reorganziation
    * ...

2. **prepare** - *writing all resources to disk*
    * putting a finalizier on the NS to prevent its deletion on vintage
    * writing all `apps` to disk
    * writing all dependend `cm`/`secrets` to disk
    * converting vintage `cm`/`secrets` locations to capi org-namespace

*...[Infrastructure migration](https://github.com/giantswarm/capi-migration-cli) should happen here...*

3. **apply** - *applying the resources to the new MC*
    * checking if certain default resources are available
    * applying the dumped resources to the new MC
    * removing the finalizier on the vintage NS

## Recomendation to run the tool
* To ensure there are no interference with kubeconfigs that the tool uses, create a new temporary file for kubeconfig.

```
❯❯❯ export KUBECONFIG=$(mktemp)
                                <sourceMC>  <destMC> <WC Name> <Org Name>
❯❯❯ ./app-migration-cli prepare -s gaia     -d golem -n ulli30 -o org-ulli
Connected to gs-gaia, k8s server version v1.24.17
Connected to gs-golem, k8s server version v1.24.16
Finalizer set on NS: gaia-ulli28
Scheduled 1 non-default apps for migration

❯❯❯ ./capi-migration-cli --mc-capi golem --mc-vintage gaia --cluster-namespace org-ulli --cluster-name ulli30
[...]
Deleted vintage ulli30 node pool ASG.
Finished migrating cluster ulli30 to CAPI infrastructure

❯❯❯ ./app-migration-cli apply -s gaia -d golem -n ulli30 -f ulli30-apps.yaml -o org-ulli
Connected to gs-gaia, k8s server version v1.24.17
Connected to gs-golem, k8s server version v1.24.16

All prerequistes are found on the new MC for app migration
Applying all non-default APP CRs to MC
All non-default apps applied successfully.

Finalizer removed on NS: gaia-ulli30
```

## Notes 
* currently only working for vintage
* currently only working for aws based clusters
