# Kubecraft

Kubecraft is a Kubernetes controller used to deploy and managed Minecraft Bedrock servers. Servers are deployed as Statefulsets in order to support persistent voumes with both Read-Write-Once and Read-Write-Many storage modes. 


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
Minecraft servers deployed with Kubecraft store the Minecraft installation and world data in a minecraft-data volume. By default, the volumn is mounted to pods as a emptyDir, which are ephemeral and do not persist beyond the lifecycle of a pod. 

Persistent Disks are essential to persisting world data. The `dataVolume` field is used for configuring
persistent volumes for your Minecraft server's installation and world data.

**DataVolume**
| Key | Description |
| -- | -- |
| `volume` | Configure a volume of type emptyDir (default) or persistentVolumeClaim for storing installation files and world data |
| `persistentVolumeClaim` | Dynamically generated a PersistentVolumeClaim configuration for your server's world data. Configuration will be added to the statefulsets PersistentVolumeClaimTemplates object. |

**Volume**
| Key | Description |
| --- | ----------- |
| emtpyDir | Create an empty diretory on the host and mounted it. This volume type is not persistent |
| persistentVolumeClaim | Configure a volume to reference an existing PVC and mount it. |


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

PersistentVolumeClaimTemplates are used with Statefulsets in order to dynamically create PersistentVolumeClaims rather than manually create them. As MinecraftServer resources are used to create and manage Statefulsets, they they support the use of PersistentVolumeClaims.

The following MinecraftServer config example adds a PersistentVolumnClaim configuration, which will be used to configure a Statefulset with a PersistentVolumeClaimTemplate entry. 


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

## Statefulset
The Kubecraft controller deploys a Statefulset API object based on the configurations of a MinecraftServer custom resource object. Using the `public-server` example above, the following statefulset configuration would be created by the controller.

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  creationTimestamp: "2024-02-28T05:36:50Z"
  generation: 1
  name: public-server
  namespace: default
  ownerReferences:
  - apiVersion: kubecraft.ca/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: MinecraftServer
    name: public-server
    uid: e5c66a6e-601c-43e6-a02d-5df4c0c10aaa
  resourceVersion: "388855"
  uid: ab2c100b-990d-4803-8ef2-6aa4385b6fa4
spec:
  persistentVolumeClaimRetentionPolicy:
    whenDeleted: Retain
    whenScaled: Retain
  podManagementPolicy: OrderedReady
  replicas: 2
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: public-server
  serviceName: ""
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: public-server
    spec:
      containers:
      - command:
        - /usr/local/bin/box64
        - /data/bedrock_server
        image: kubecraft/minecraft-server:v0.1.0
        imagePullPolicy: IfNotPresent
        name: minecraft
        resources:
          limits:
            cpu: "2"
            memory: 4Gi
          requests:
            cpu: 100m
            memory: 696Mi
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /data
          name: minecraft-data
      dnsPolicy: ClusterFirst
      initContainers:
      - args:
        - --version
        - 1.20.62.02
        - --output
        - /data
        command:
        - /kubecraft-cli
        image: kubecraft/kubecraft-cli:v0.1.0
        imagePullPolicy: IfNotPresent
        name: init
        resources: {}
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /data
          name: minecraft-data
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      serviceAccount: minecraft-server
      serviceAccountName: minecraft-server
      terminationGracePeriodSeconds: 30
      volumes:
      - name: minecraft-data
        persistentVolumeClaim:
          claimName: public-server-mcdata
  updateStrategy:
    rollingUpdate:
      partition: 0
    type: RollingUpdate
status:
  availableReplicas: 2
  collisionCount: 0
  currentReplicas: 2
  currentRevision: public-server-84b75855c9
  observedGeneration: 1
  readyReplicas: 2
  replicas: 2
  updateRevision: public-server-84b75855c9
  updatedReplicas: 2
```