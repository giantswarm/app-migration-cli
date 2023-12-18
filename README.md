# app-migration-cli

This tool migrates `apps` from one MC to another. It's closly coupled
to the vintage -> capi [migration tool](https://github.com/giantswarm/capi-migration-cli).

> :warning: WIP

## :airplane: Rundown

Parts of the script must run in advance of the CAPI Migration, others
are scheduled after a successfull infrastructure migration.

1. preflight - *readonly checks if a migration is possible*
    * validate access to both mcs
    * check WC condition/health (only "Created" is allowed)

2. prepare - *writing all resources to disk*
    * putting a finalizier on the NS to prevent its deletion
    * writing all `apps` to disk
    * writing all dependend `cm`/`secrets` to disk

*...Infrastructure migration should happen here...*

3. apply - *applying the resources to the new MC*
    * applying the dumped resources to the new MC
    * removing the ns/finalizier

## Recomendation to run the tool
To ensure there are no interference with kubeconfigs that the tool uses, create a new temporary file for kubeconfig.
```
export KUBECONFIG=$(mktemp)
chmod 600 $KUBECONFIG

./foo bar barfoo
```

## Notes 
* currently only working for vintage
* currently only working for aws based clusters
