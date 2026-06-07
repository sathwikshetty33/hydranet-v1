export const PIPELINE = [
  {
    id: '01',
    title: 'Trip the wire',
    technique: 'WEBHOOK INGEST',
    body: 'Your monitor — Falco, Datadog, anything — fires a webhook the instant it sees a breach. HydraNet catches it and opens an incident.',
  },
  {
    id: '02',
    title: 'Seal the room',
    technique: 'IN-PLACE ISOLATION',
    body: 'A NetworkPolicy drops over the pod. Lateral movement dies; outbound stays open. The attacker never notices the walls go up.',
  },
  {
    id: '03',
    title: 'Mark the loot',
    technique: 'SECRET ENUMERATION',
    body: 'Every mounted secret, token, and config the pod could touch is catalogued as exposed and pushed to your rotation endpoint.',
  },
  {
    id: '04',
    title: 'Watch them work',
    technique: 'LIVE FORENSICS',
    body: 'Tetragon syscalls and Hubble flows stream to two LLM agents — one reconstructs the break-in, one narrates intent in real time.',
  },
  {
    id: '05',
    title: 'Pull the plug',
    technique: 'ESCAPE INTERLOCK',
    body: 'The moment they probe for a container escape, the pod is terminated. Intelligence is already captured. The host stays clean.',
  },
];

export const CAPABILITIES = [
  {
    k: 'Containment',
    h: 'Trap, don’t kill',
    p: 'Most playbooks nuke a compromised pod on sight and torch the evidence. HydraNet keeps the session breathing inside a sealed cell.',
  },
  {
    k: 'Attribution',
    h: 'MITRE ATT&CK chains',
    p: 'The forensics agent reads the 15 minutes before the alarm and maps initial access to a technique ID with supporting events.',
  },
  {
    k: 'Surveillance',
    h: 'Intent, live',
    p: 'A fast model classifies behavior every 30 seconds — recon, credential theft, exfil, mining, beaconing — as it happens.',
  },
  {
    k: 'Reflex',
    h: 'Escape interlock',
    p: 'Capability probes, docker.sock, host mounts, nsenter. The state machine escalates and terminates before the exploit lands.',
  },
];

export const MODES = [
  {
    name: 'ENHANCED',
    cni: 'Cilium detected',
    detail: 'CiliumNetworkPolicy with L7 rules. Full Hubble flow visibility — DNS, HTTP, TCP.',
  },
  {
    name: 'FULL',
    cni: 'Calico / vanilla',
    detail: 'L3/L4 NetworkPolicy isolation with complete syscall-level monitoring via Tetragon.',
  },
  {
    name: 'LIMITED',
    cni: 'No policy CNI',
    detail: 'Forensics and behavior monitoring only. No active isolation. A loud warning is logged.',
  },
];

export const STACK = [
  ['Go core', 'Gin · client-go · controller-runtime'],
  ['Analysis', 'Python · gRPC · Anthropic'],
  ['Observe', 'Tetragon (eBPF) · Hubble'],
  ['Store', 'Postgres'],
];
