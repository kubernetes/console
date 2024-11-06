// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import {KdError, StringMap} from '@api/root.shared';
import {PersistentVolumeSource} from '@api/volume.api';

export interface TypeMeta {
  kind: string;
  scalable?: boolean;
  restartable?: boolean;
}

export interface ListMeta {
  totalItems: number;
}

export interface ObjectMeta {
  name?: string;
  namespace?: string;
  labels?: StringMap;
  annotations?: StringMap;
  creationTimestamp?: string;
  uid?: string;
  ownerReferences?: Array<OwnerReference>;
}

export interface OwnerReference {
  kind?: string;
  name?: string;
}

export interface JobStatus {
  status: string;
  message: string;
  conditions: Condition[];
}

export interface ResourceDetail {
  objectMeta: ObjectMeta;
  typeMeta: TypeMeta;
  errors: K8sError[];
}

export interface ResourceList {
  listMeta: ListMeta;
  items?: Resource[];
  errors?: K8sError[];
}

export interface Resource {
  objectMeta: ObjectMeta;
  typeMeta: TypeMeta;
}

export interface ResourceOwner extends Resource {
  pods: PodInfo;
  containerImages: string[];
  initContainerImages: string[];
}

export interface LabelSelector {
  matchLabels: StringMap;
}

export interface CapacityItem {
  resourceName: string;
  quantity: string;
}

// List types
export interface ClusterRoleList extends ResourceList {
  items: ClusterRole[];
}

export interface ClusterRoleBindingList extends ResourceList {
  items: ClusterRoleBinding[];
}

export interface RoleList extends ResourceList {
  items: Role[];
}

export interface RoleBindingList extends ResourceList {
  items: RoleBinding[];
}

export interface ConfigMapList extends ResourceList {
  items: ConfigMap[];
}

export interface CronJobList extends ResourceList {
  cumulativeMetrics: Metric[] | null;
  items: CronJob[];
  status: Status;
}

export interface CRDList extends ResourceList {
  items: CRD[];
}

export interface CRDObjectList extends ResourceList {
  typeMeta: TypeMeta;
  items: CRDObject[];
}

export interface DaemonSetList extends ResourceList {
  cumulativeMetrics: Metric[] | null;
  daemonSets: DaemonSet[];
  status: Status;
}

export interface DeploymentList extends ResourceList {
  cumulativeMetrics: Metric[] | null;
  deployments: Deployment[];
  status: Status;
}

export interface EndpointList extends ResourceList {
  endpoints: Endpoint[];
}

export interface EventList extends ResourceList {
  events: Event[];
}

export interface HorizontalPodAutoscalerList extends ResourceList {
  horizontalpodautoscalers: HorizontalPodAutoscaler[];
}

export interface IngressList extends ResourceList {
  items: Ingress[];
}

export interface ServiceAccountList extends ResourceList {
  items: ServiceAccount[];
}

export interface NetworkPolicyList extends ResourceList {
  items: NetworkPolicy[];
}

export interface JobList extends ResourceList {
  cumulativeMetrics: Metric[] | null;
  jobs: Job[];
  status: Status;
}

export interface NamespaceList extends ResourceList {
  namespaces: Namespace[];
}

export interface NodeList extends ResourceList {
  cumulativeMetrics: Metric[] | null;
  nodes: Node[];
}

export interface PersistentVolumeClaimList extends ResourceList {
  items: PersistentVolumeClaim[];
}

export interface PersistentVolumeList extends ResourceList {
  items: PersistentVolume[];
}

export interface PodContainerList {
  containers: string[];
}

export interface PodList extends ResourceList {
  pods: Pod[];
  status: Status;
  podInfo?: PodInfo;
  cumulativeMetrics: Metric[] | null;
}

export interface ReplicaSetList extends ResourceList {
  cumulativeMetrics: Metric[] | null;
  replicaSets: ReplicaSet[];
  status: Status;
}

export interface ReplicationControllerList extends ResourceList {
  cumulativeMetrics: Metric[] | null;
  replicationControllers: ReplicationController[];
  status: Status;
}

export interface ResourceQuotaDetailList extends ResourceList {
  items: ResourceQuotaDetail[];
}

export interface SecretList extends ResourceList {
  secrets: Secret[];
}

export interface ServiceList extends ResourceList {
  services: Service[];
}

export interface StatefulSetList extends ResourceList {
  cumulativeMetrics: Metric[] | null;
  statefulSets: StatefulSet[];
  status: Status;
}

export interface StorageClassList extends ResourceList {
  items: StorageClass[];
}

export interface IngressClassList extends ResourceList {
  items: IngressClass[];
}

// Simple detail types
export type ClusterRole = Resource;

export type ClusterRoleBinding = Resource;

export type Role = Resource;

export type RoleBinding = Resource;

export type ConfigMap = Resource;

export type ServiceAccount = Resource;

export type NetworkPolicy = Resource;

export interface Controller extends Resource {
  pods: PodInfo;
  containerImages: string[];
  initContainerImages: string[];
}

export interface CronJob extends Resource {
  schedule: string;
  suspend: boolean;
  active: number;
  lastSchedule: string;
  containerImages: string[];
}

export interface CRD extends Resource {
  group: string;
  scope: string;
  nameKind: string;
  established: string;
}

export type CRDObject = Resource;

export interface DaemonSet extends Resource {
  podInfo: PodInfo;
  containerImages: string[];
  initContainerImages: string[];
}

export interface Deployment extends Resource {
  pods: PodInfo;
  containerImages: string[];
  initContainerImages: string[];
}

export interface EndpointResourceList extends ResourceList {
  endpoints: EndpointResource[];
}

export interface EndpointResource extends Resource {
  host: string;
  nodeName: string;
  ready: boolean;
  ports: EndpointResourcePort[];
}

export interface EndpointResourcePort {
  name: string;
  port: number;
  protocol: string;
}

export interface Port {
  port: number;
  name: string;
  protocol: string;
  nodePort?: number;
}

export interface Endpoint {
  host: string;
  nodeName?: string;
  ports: Port[];
  ready?: boolean;
  typeMeta?: TypeMeta;
  objectMeta?: ObjectMeta;
}

export interface Event extends Resource {
  message: string;
  sourceComponent: string;
  sourceHost: string;
  object: string;
  objectKind?: string;
  objectName?: string;
  objectNamespace?: string;
  count: number;
  firstSeen: string;
  lastSeen: string;
  reason: string;
  type: string;
}

export interface HorizontalPodAutoscaler extends Resource {
  scaleTargetRef: ScaleTargetRef;
  minReplicas: number;
  maxReplicas: number;
  currentCPUUtilization: number;
  targetCPUUtilization?: number;
}

export interface Ingress extends Resource {
  endpoints: Endpoint[];
  hosts: string[];
}

export interface Job extends Resource {
  podInfo: PodInfo;
  containerImages: string[];
  initContainerImages: string[];
  parallelism: number;
}

export interface Namespace extends Resource {
  phase: string;
}

export interface Node extends Resource {
  ready: string;
}

export interface PersistentVolume extends Resource {
  capacity: StringMap;
  storageClass: string;
  accessModes: string[];
  reclaimPolicy: string;
  mountOptions?: string[];
  status: string;
  claim: string;
  reason: string;
}

export interface PersistentVolumeClaim extends Resource {
  status: string;
  volume: string;
}

export interface Pod extends Resource {
  status: string;
  podIP?: string;
  restartCount: number;
  qosClass?: string;
  metrics: PodMetrics;
  warnings: Event[];
  nodeName: string;
  serviceAccountName: string;
  containerImages: string[];
}

export interface PodContainer {
  name: string;
  restartCount: number;
}

export interface ReplicaSet extends Resource {
  podInfo: PodInfo;
  containerImages: string[];
  initContainerImages: string[];
}

export interface ReplicationController extends Resource {
  podInfo: PodInfo;
  containerImages: string[];
  initContainerImages: string[];
}

export interface Secret extends Resource {
  type: string;
}

export interface Service extends Resource {
  internalEndpoint: Endpoint;
  externalEndpoints: Endpoint[];
  selector: StringMap;
  type: string;
  clusterIP: string;
}

export interface StatefulSet extends Resource {
  podInfo: PodInfo;
  containerImages: string[];
  initContainerImages: string[];
}

export interface StorageClass extends Resource {
  provisioner: string;
  parameters: StringMap[];
}

export interface IngressClass extends Resource {
  controller: string;
}

// Detail types

export interface ReplicaSetDetail extends ResourceDetail {
  selector: LabelSelector;
  podInfo: PodInfo;
  podList: PodList;
  containerImages: string[];
  initContainerImages: string[];
  eventList: EventList;
}

export interface ResourceQuotaDetail extends ResourceDetail {
  scopes: string[];
  statusList: {[key: string]: ResourceQuotaStatus};
}

export interface DeploymentDetail extends ResourceDetail {
  selector: Label[];
  statusInfo: DeploymentInfo;
  conditions: Condition[];
  strategy: string;
  minReadySeconds: number;
  revisionHistoryLimit?: number;
  rollingUpdateStrategy?: RollingUpdateStrategy;
  events: EventList;
}

export interface ReplicationControllerDetail extends ResourceDetail {
  labelSelector: StringMap;
  containerImages: string[];
  initContainerImages: string[];
  podInfo: PodInfo;
  podList: PodList;
  serviceList: ServiceList;
  eventList: EventList;
  hasMetrics: boolean;
}

export interface ServiceDetail extends ResourceDetail {
  internalEndpoint: Endpoint;
  externalEndpoints: Endpoint[];
  endpointList: EndpointList;
  selector: StringMap;
  type: string;
  clusterIP: string;
  podList: PodList;
  sessionAffinity: string;
}

export interface DaemonSetDetail extends ResourceDetail {
  labelSelector: StringMap;
  containerImages: string[];
  initContainerImages: string[];
  podInfo: PodInfo;
}

export interface NamespaceDetail extends ResourceDetail {
  phase: string;
  eventList: EventList;
  resourceLimits: LimitRange[];
  resourceQuotaList: ResourceQuotaDetailList;
}

export interface PolicyRule {
  verbs: string[];
  apiGroups: string[];
  resources: string[];
  resourceNames: string[];
  nonResourceURLs: string[];
}

export interface ClusterRoleDetail extends ResourceDetail {
  rules: PolicyRule[];
}

export interface Subject {
  kind: string;
  apiGroup: string;
  name: string;
  namespace: string;
}

export interface ResourceRef {
  kind: string;
  apiGroup: string;
  name: string;
}

export interface ClusterRoleBindingDetail extends ResourceDetail {
  subjects: Subject[];
  roleRef: ResourceRef;
}

export interface RoleDetail extends ResourceDetail {
  rules: PolicyRule[];
}

export interface RoleBindingDetail extends ResourceDetail {
  subjects: Subject[];
  roleRef: ResourceRef;
}

export interface SecretDetail extends ResourceDetail {
  type: string;
  data: StringMap;
}

export type ServiceAccountDetail = ResourceDetail;

export interface IngressDetail extends ResourceDetail {
  endpoints: Endpoint[];
  spec: IngressSpec;
}

export interface IngressSpec {
  ingressClassName?: string;
  defaultBackend?: IngressBackend;
  rules?: IngressSpecRule[];
  tls?: IngressSpecTLS[];
}

export interface IngressSpecTLS {
  hosts: string[];
  secretName: string;
}

export interface IngressBackend {
  service?: IngressBackendService;
  resource?: ResourceRef;
}

export interface IngressBackendService {
  name: string;
  port: IngressBackendServicePort;
}

export interface IngressBackendServicePort {
  name?: string;
  number?: number;
}

export interface IngressSpecRule {
  host?: string;
  http: IngressSpecRuleHttp;
}

export interface IngressSpecRuleHttp {
  paths: IngressSpecRuleHttpPath[];
}

export interface IngressSpecRuleHttpPath {
  path: string;
  pathType: string;
  backend: IngressBackend;
}

export interface NetworkPolicyDetail extends ResourceDetail {
  podSelector: LabelSelector;
  ingress?: any;
  egress?: any;
  policyTypes?: string[];
}

export interface PersistentVolumeClaimDetail extends ResourceDetail {
  status: string;
  volume: string;
  capacity: string;
  storageClass: string;
  accessModes: string[];
}

export interface StorageClassDetail extends ResourceDetail {
  parameters: StringMap;
  provisioner: string;
}

export interface IngressClassDetail extends ResourceDetail {
  parameters: StringMap;
  controller: string;
}

export interface ConfigMapDetail extends ResourceDetail {
  data: StringMap;
}

export interface CRDDetail extends ResourceDetail {
  version?: string;
  group: string;
  scope: string;
  names: CRDNames;
  versions: CRDVersion[];
  objects: CRDObjectList;
  conditions: Condition[];
  subresources: string[];
}

export type CRDObjectDetail = ResourceDetail;

export interface JobDetail extends ResourceDetail {
  podInfo: PodInfo;
  podList: PodList;
  containerImages: string[];
  initContainerImages: string[];
  eventList: EventList;
  parallelism: number;
  completions: number;
  jobStatus: JobStatus;
}

export interface CronJobDetail extends ResourceDetail {
  schedule: string;
  suspend: boolean;
  active: number;
  lastSchedule: string;
  concurrencyPolicy: string;
  startingDeadlineSeconds: number;
}

export interface StatefulSetDetail extends ResourceDetail {
  podInfo: PodInfo;
  podList: PodList;
  containerImages: string[];
  initContainerImages: string[];
  eventList: EventList;
}

export interface PersistentVolumeDetail extends ResourceDetail {
  status: string;
  claim: string;
  reclaimPolicy: string;
  accessModes: string[];
  capacity: StringMap;
  message: string;
  storageClass: string;
  reason: string;
  persistentVolumeSource: PersistentVolumeSource;
  mountOptions?: string[];
}

export interface PodDetail extends ResourceDetail {
  initContainers: Container[];
  containers: Container[];
  podPhase: string;
  podIP: string;
  nodeName: string;
  restartCount: number;
  qosClass: string;
  metrics: Metric[];
  conditions: Condition[];
  controller: Resource;
  imagePullSecrets: LocalObjectReference[];
  eventList: EventList;
  persistentVolumeClaimList: PersistentVolumeClaimList;
  securityContext: PodSecurityContext;
}

export interface LocalObjectReference {
  name: string;
}

export interface NodeDetail extends ResourceDetail {
  phase: string;
  podCIDR: string;
  providerID: string;
  unschedulable: boolean;
  allocatedResources: NodeAllocatedResources;
  nodeInfo: NodeInfo;
  containerImages: string[];
  initContainerImages: string[];
  addresses: NodeAddress[];
  taints: NodeTaint[];
  metrics: Metric[];
  conditions: Condition[];
  podList: PodList;
  eventList: EventList;
}

export interface HorizontalPodAutoscalerDetail extends ResourceDetail {
  scaleTargetRef: ScaleTargetRef;
  minReplicas: number;
  maxReplicas: number;
  currentCPUUtilization: number;
  targetCPUUtilization?: number;
  currentReplicas: number;
  desiredReplicas: number;
  lastScaleTime: string;
}

// Validation types
export interface AppNameValidity {
  valid: boolean;
}

export interface AppNameValiditySpec {
  name: string;
  namespace: string;
}

export interface ImageReferenceValidity {
  valid: boolean;
  reason: string;
}

export interface ImageReferenceValiditySpec {
  reference: string;
}

export interface ProtocolValidity {
  valid: boolean;
}

export interface ProtocolValiditySpec {
  protocol: string;
  isExternal: boolean;
}

// Auth related types
export interface AuthResponse {
  token: string;
}

export interface CanIResponse {
  allowed: boolean;
}

export interface AppDeploymentContentSpec {
  name: string;
  namespace: string;
  content: string;
  validate: boolean;
}

export interface AppDeploymentContentResponse {
  error: string;
  contet: string;
  name: string;
}

export interface AppDeploymentSpec {
  containerImage: string;
  containerCommand?: string;
  containerCommandArgs?: string;
  isExternal: boolean;
  name: string;
  description?: string;
  portMappings: PortMapping[];
  labels: Label[];
  replicas: number;
  namespace: string;
  memoryRequirement?: string;
  cpuRequirement?: number;
  runAsPrivileged: boolean;
  imagePullSecret: string;
  variables: EnvironmentVariable[];
}

export interface CsrfToken {
  token: string;
}

export interface LocalSettings {
  theme: string;
  timezone: string;
}

export interface Theme {
  name: string;
  displayName: string;
  isDark: boolean;
}

export interface AppConfig {
  serverTime: number;
  userAgent: string;
  version: string;
}

export interface ErrStatus {
  message: string;
  code: number;
  status: string;
  reason: string;
}

export interface K8sError {
  ErrStatus: ErrStatus;

  toKdError(): KdError;
}

export interface Condition {
  type: string;
  status: string;
  lastProbeTime: string;
  lastTransitionTime: string;
  reason: string;
  message: string;
}

export interface ContainerStateWaiting {
  reason?: string;
  message?: string;
}

export interface ContainerStateRunning {
  startedAt?: string;
}

export interface ContainerStateTerminated {
  exitCode: number;
  reason?: string;
  message?: string;
  signal?: number;
}

export interface ContainerState {
  waiting?: ContainerStateWaiting;
  terminated?: ContainerStateTerminated;
  running?: ContainerStateRunning;
}

export interface ResourceQuotaStatus {
  used: string;
  hard: string;
}

export interface MetricResult {
  timestamp: string;
  value: number;
}

export interface Metric {
  dataPoints: DataPoint[];
  metricName: string;
  aggregation: string;
}

export interface DataPoint {
  x: number;
  y: number;
}

export interface ConfigMapKeyRef {
  name: string;
  key: string;
}

export interface SecretKeyRef {
  name: string;
  key: string;
}

export interface EnvVar {
  name: string;
  value: string;
  valueFrom: EnvVarSource;
}

export interface EnvVarSource {
  configMapKeyRef: ConfigMapKeyRef;
  secretKeyRef: SecretKeyRef;
}

export interface Container {
  name: string;
  image: string;
  env: EnvVar[];
  commands: string[];
  args: string[];
  volumeMounts: VolumeMounts[];
  securityContext: ContainerSecurityContext;
  status: ContainerStatus;
  livenessProbe: Probe;
  readinessProbe: Probe;
  startupProbe: Probe;
  resources?: ResourceRequirements;
}

export interface ResourceRequirements {
  limits?: StringMap;
  requests?: StringMap;
}

export interface Probe {
  httpGet?: ProbeHttpGet;
  tcpSocket?: ProbeTcpSocket;
  exec?: ProbeExec;
  initialDelaySeconds?: number;
  timeoutSeconds?: number;
  periodSeconds?: number;
  successThreshold?: number;
  failureThreshold?: number;
  terminationGracePeriodSeconds?: number;
}

export interface ProbeHttpGet {
  path?: string;
  port: string | number;
  host?: string;
  scheme?: string;
  httpHeaders?: string[];
}

export interface ProbeTcpSocket {
  port: string | number;
  host?: string;
}

export interface ProbeExec {
  command?: string[];
}

export interface ContainerStatus {
  name: string;
  state: ContainerState;
  lastTerminationState: ContainerState;
  ready: boolean;
  restartCount: number;
  started?: boolean;
}

export interface ISecurityContext {
  seLinuxOptions?: SELinuxOptions;
  windowsOptions?: WindowsSecurityContextOptions;
  runAsUser?: number;
  runAsGroup?: number;
  runAsNonRoot?: boolean;
  seccompProfile?: SeccompProfile;
}

export interface ContainerSecurityContext extends ISecurityContext {
  capabilities?: Capabilities;
  privileged?: boolean;
  readOnlyRootFilesystem?: boolean;
  allowPrivilegeEscalation?: boolean;
  procMount?: string; // ProcMountType;
}

export interface PodSecurityContext extends ISecurityContext {
  fsGroup?: number;
  fsGroupChangePolicy?: string;
  supplementalGroups?: number[];
  sysctls?: Sysctl[];
}

export interface Sysctl {
  name: string;
  value: string;
}

export interface Capabilities {
  add: string[];
  drop: string[];
}

export interface SELinuxOptions {
  user?: string;
  role?: string;
  type?: string;
  level?: string;
}

export interface WindowsSecurityContextOptions {
  gMSACredentialSpecName?: string;
  gMSACredentialSpec?: string;
  runAsUserName?: string;
}

export interface SeccompProfile {
  type: string; // SeccompProfileType;
  localhostProfile?: string;
}

export interface VolumeMounts {
  name: string;
  readOnly: boolean;
  mountPath: string;
  subPath: string;
  volume: PersistentVolumeSource;
}

export interface CRDNames {
  plural: string;
  singular?: string;
  shortNames?: string[];
  kind: string;
  listKind?: string;
  categories?: string[];
}

export interface CRDVersion {
  name: string;
  served: boolean;
  storage: boolean;
}

export interface PodMetrics {
  cpuUsage: number;
  memoryUsage: number;
  cpuUsageHistory: MetricResult[];
  memoryUsageHistory: MetricResult[];
}

export interface Status {
  running: number;
  failed: number;
  pending: number;
  succeeded: number;
}

export interface PodStatus {
  podPhase: string;
  status: string;
  containerStates: ContainerState[];
}

export interface PodInfo {
  current: number;
  desired: number;
  running: number;
  pending: number;
  failed: number;
  succeeded: number;
  warnings: Event[];
}

export interface NodeAllocatedResources {
  cpuRequests: number;
  cpuRequestsFraction: number;
  cpuLimits: number;
  cpuLimitsFraction: number;
  cpuCapacity: number;
  memoryRequests: number;
  memoryRequestsFraction: number;
  memoryLimits: number;
  memoryLimitsFraction: number;
  memoryCapacity: number;
  allocatedPods: number;
  podCapacity: number;
  podFraction: number;
}

export interface NodeInfo {
  machineID: string;
  systemUUID: string;
  bootID: string;
  kernelVersion: string;
  osImage: string;
  containerRuntimeVersion: string;
  kubeletVersion: string;
  kubeProxyVersion: string;
  operatingSystem: string;
  architecture: string;
}

export interface NodeAddress {
  type: string;
  address: string;
}

export interface NodeTaint {
  key: string;
  value: string;
  effect: string;
  timeAdded: number;
}

export interface PortMapping {
  port: number | null;
  protocol: string;
  targetPort: number | null;
}

export interface EnvironmentVariable {
  name: string;
  value: string;
}

export interface Label {
  key: string;
  value: string;
}

export interface PodEvent {
  reason: string;
  message: string;
}

export interface RollingUpdateStrategy {
  maxSurge: number | string;
  maxUnavailable: number | string;
}

export interface DeploymentInfo {
  replicas: number;
  updated: number;
  available: number;
  unavailable: number;
}

export interface ReplicationControllerSpec {
  replicas: number;
}

export interface ReplicaCounts {
  desiredReplicas: number;
  actualReplicas: number;
}

export interface DeleteReplicationControllerSpec {
  deleteServices: boolean;
}

export interface NamespaceSpec {
  name: string;
}

export interface ReplicationControllerPodWithContainers {
  name: string;
  startTime?: string;
  totalRestartCount: number;
  podContainers: PodContainer[];
}

export interface ReplicationControllerPods {
  pods: ReplicationControllerPodWithContainers[];
}

export interface LogSources {
  podNames: string[];
  containerNames: string[];
  initContainerNames: string[];
}

export interface LogDetails {
  info: LogInfo;
  logs: LogLine[];
  selection: LogSelection;
}

export interface LogInfo {
  podName: string;
  containerName: string;
  initContainerName: string;
  fromDate: string;
  toDate: string;
  truncated: boolean;
}

export interface LogLine {
  timestamp: string;
  content: string;
}

export enum LogControl {
  LoadStart = 'beginning',
  LoadEnd = 'end',
  TimestampOldest = 'oldest',
  TimestampNewest = 'newest',
}

export interface LogSelection {
  logFilePosition: LogControl;
  referencePoint: LogLineReference;
  offsetFrom: number;
  offsetTo: number;
}

export interface LogLineReference {
  timestamp: LogControl;
  lineNum: number;
}

export type LogOptions = {
  previous: boolean;
  timestamps: boolean;
};

export interface Protocols {
  protocols: string[];
}

export interface SecretSpec {
  name: string;
  namespace: string;
  data: string;
}

export interface LimitRange {
  resourceType: string;
  resourceName: string;
  min: string;
  max: string;
  default: string;
  defaultRequest: string;
  maxLimitRequestRatio: string;
}

export interface ScaleTargetRef {
  kind: string;
  name: string;
}

export interface GlobalSettings {
  clusterName: string;
  itemsPerPage: number;
  labelsLimit: number;
  logsAutoRefreshTimeInterval: number;
  resourceAutoRefreshTimeInterval: number;
  disableAccessDeniedNotifications: boolean;
  hideAllNamespaces: boolean;
  defaultNamespace: string;
  namespaceFallbackList: string[];
}

export interface PinnedResource {
  kind: string;
  name: string;
  displayName: string;
  namespace?: string;
  namespaced: boolean;
}

export interface APIVersion {
  name: string;
}

export interface LoginSpec {
  username: string;
  password: string;
  token: string;
  kubeConfig: string;
}

export type AuthenticationMode = string;

export interface EnabledAuthenticationModes {
  modes: AuthenticationMode[];
}

export interface LoginSkippableResponse {
  skippable: boolean;
}

export interface SystemBanner {
  message: string;
  severity: string;
}

export interface TerminalResponse {
  id: string;
}

export interface ShellFrame {
  Op: string;
  Data?: string;
  SessionID?: string;
  Rows?: number;
  Cols?: number;
}

export interface TerminalPageParams {
  namespace: string;
  resourceKind: string;
  resourceName: string;
  pod?: string;
  container?: string;
}

export interface SockJSSimpleEvent {
  type: string;

  toString(): string;
}

export interface SJSCloseEvent extends SockJSSimpleEvent {
  code: number;
  reason: string;
  wasClean: boolean;
}

export interface SJSMessageEvent extends SockJSSimpleEvent {
  data: string;
}

export interface User {
  name: string;
  authenticated: boolean;
}

export interface Timezone {
  value: string;
  label: string;
}
