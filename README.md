# app-migration-cli

This tool migrates `apps` from one MC to another


### Recomendation to run the tool
To ensure there are no interference with kubeconfigs that the tool uses, create a new temporary file for kubeconfig.
```
export KUBECONFIG=$(mktemp)
chmod 600 $KUBECONFIG
```

