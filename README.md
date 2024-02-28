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

- **serverVersion**: sets the version of Minecraft Server to download
- **dataVolume**: To preserve world data persistent volumes are used. This field can be used to set either a Persistent Volume Claim or a PersistentVolumeClaimTemplate.
- **serverSettings**: Is where you set the properties for the server.properties file.

## Persistent Data
Persistent Disks are essential to persisting world data beyond the life of a pod. The `dataVolume` field is used for configuring
persistent volumes for your Minecraft server's installation and world data.

| `persistentVolumeClaim` | Reference an existing Persistent Volume Claim configuration on the server. This is the recommended setting for running multiple replicas of a server that share the same world data. |
| `persistentVolumeClaimtemplate` | Dynamically generated a PersistentVolumeClaim configuration for your server's world data |


### Using PersistentVolumeClaim

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

### Using PersistentVolumeClaimTemplate

```yaml
apiVersion: kubecraft.ca/v1alpha1
kind: MinecraftServer
metadata:
  name: private-server-a
spec:
  serverVersion: "1.20.62.02"
  createService: true
  replicas: 1
  serviceAccountName: "minecraft-server"
  dataVolume:
    persistentVolumeClaim:
      storageClassName: standard
      accessModes:
      - ReadWriteOncePod
      resources:
        requests:
          storage: 1Gi
      volumeMode: Filesystem
  serverSettings:
    gameMode: "survival"
    levelSeed: "default"
```

