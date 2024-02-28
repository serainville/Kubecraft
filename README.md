# Kubecraft

Kubecraft is a Kubernetes controller used to deploy and managed Minecraft Bedrock servers. 

## MinecraftServer

MinecraftServer is a customer resource definition (CRD) used to generate a statefulset for your
Minecraft servers.

```yaml
apiVersion: kubecraft.ca/v1alpha1
kind: MinecraftServer
metadata:
  name: public-server
spec:
  serverVersion: "1.20.62.02"
  createService: true
  replicas: 2
  serviceAccountName: "minecraft-server"
  dataVolume:
    volume:
      persistentVolumeClaim:
        claimName:  public-server-mcdata
  serverSettings:
    gameMode: "survival"
    levelSeed: "default"
```

- *serverVersion*: sets the version of Minecraft Server to download
- *dataVolume: To preserve world data persistent volumes are used. This field can be used to set either a Persistent Volume Claim or a PersistentVolumeClaimTemplate.
- *serverSettings*: Is where you set the properties for the server.properties file.



