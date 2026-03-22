# Agent Visualizer — Design Proposal

**Status:** Draft
**Location:** `extras/agent-viz/`

## Overview

A standalone Three.js-based 3D graph visualization tool that renders real-time agent activity within a Scion grove. The visualizer connects to the Hub's existing SSE event stream and REST API to display:

- **Agent lifecycle** — nodes appear/disappear as agents are created and destroyed
- **Messaging** — animated particles flow between nodes when messages are sent (agent-to-agent, human-to-agent, system-to-agent)
- **State changes** — node color/shape/glow shifts to reflect agent phase and activity
- **File activity** — visual indicators when agents create or modify files

## Data Sources

The Hub already exposes everything needed — no backend changes required for an MVP.

### SSE Event Stream (`GET /events`)

Subscribe with patterns like `grove.{groveId}.>` to receive:

| Event Subject | Payload | Visualizer Use |
|---|---|---|
| `grove.{id}.agent.created` | `AgentCreatedEvent` (name, slug, template, phase) | Add node to graph |
| `grove.{id}.agent.deleted` | `{agentId, groveId}` | Remove node from graph |
| `grove.{id}.agent.status` | `AgentStatusEvent` (phase, activity, detail) | Update node appearance |
| `broker.{id}.connected` | `BrokerGroveEvent` | Show broker connectivity |
| `notification.>` | Notification payloads | Overlay notification indicators |

### REST API (Initial State + Supplemental)

| Endpoint | Use |
|---|---|
| `GET /api/v1/groves/{id}` | Grove metadata for scene title/context |
| `GET /api/v1/agents?grove_id={id}` | Seed graph with existing agents on load |
| `GET /api/v1/agents/{id}` | Fetch full agent detail on click |

### Structured Logs (Stretch — Replay Mode)

The `scion-agents` and `scion-messages` log streams (as seen in `.scratch/downloaded-logs-*.json`) contain rich event data with timestamps, trace IDs, and causal chains. A replay mode could ingest exported logs and animate the visualization historically.

Key log event types available:

| Log Event | Fields | Visualizer Use |
|---|---|---|
| `agent.session.start` | agent_id, grove_id, harness | Agent becomes active |
| `agent.session.end` | agent_id | Agent session concludes |
| `agent.turn.start` / `agent.turn.end` | agent_id, session_id | Thinking/processing cycles |
| `agent.tool.call` / `agent.tool.result` | agent_id, tool_name | Tool execution (file ops visible here) |
| `agent.lifecycle.pre_start` / `post_start` / `pre_stop` | agent_id | Lifecycle transitions |
| `message dispatched` / `message accepted (buffered)` | sender, recipient, msg_type, message_content | Message flow between agents |
| `notification message dispatched` | sender, recipient, msg_type | State-change notifications |

## Visualization Design

### Node Types

| Entity | Shape | Default Color | Notes |
|---|---|---|---|
| Agent | Sphere | Based on template/harness | Primary graph element |
| Human/User | Icosahedron | White | Represents user interactions |
| System | Octahedron | Gray | Hub/broker system messages |
| Grove | Background ring/boundary | Subtle outline | Optional spatial container |

### Agent State → Visual Mapping

#### Phase (lifecycle)

| Phase | Visual |
|---|---|
| `created` | Small, translucent sphere, fading in |
| `provisioning` / `cloning` / `starting` | Pulsing/breathing animation |
| `running` | Full opacity, steady glow |
| `stopping` | Fading out animation |
| `stopped` | Dimmed, small, or removed |
| `error` | Red, jagged glow / shake animation |

#### Activity (runtime, when phase=running)

| Activity | Visual |
|---|---|
| `idle` | Steady, muted glow |
| `thinking` | Rotating ring / orbit particle |
| `executing` | Bright pulse, tool name label |
| `waiting_for_input` | Amber beacon / exclamation indicator |
| `completed` | Green check overlay, calms down |
| `blocked` | Red-amber warning pulse |
| `stalled` / `offline` | Desaturated, static |
| `limits_exceeded` | Red cap indicator |

### Message Visualization

Messages are the primary dynamic element. Using directional particles along graph edges:

- **Agent → Agent:** Colored particle matching sender, travels along link
- **Human → Agent:** Distinct particle shape (star/diamond), enters from edge of scene
- **System → Agent:** Subtle, uniform particles from a central system node
- **Broadcast:** Particle emits simultaneously to all connected links
- **Message type differentiation:**
  - `instruction` — bright, fast particle
  - `state-change` — slow, pulsing particle
  - `input-needed` — amber, attention-grabbing

### File Activity Indicators

When `agent.tool.call` events include file-related tools (`write_file`, `edit_file`, `read_file`, `run_shell_command` with file operations), show:

- Small document icons orbiting the agent node
- Brief text label with filename
- Color: green for create, blue for modify, gray for read

### Camera and Interaction

- **Auto-orbit** by default with pause on user interaction
- **Click node** → sidebar/overlay with agent detail (name, state, task summary, recent messages)
- **Click edge** → show recent messages on that channel
- **Zoom to fit** when agents are added/removed
- **Time scrubber** (replay mode) to scan through historical log data

## Technology Approach

### Recommended: `3d-force-graph` (vanilla, not React wrapper)

The scion web frontend uses **Lit** (Web Components), not React. The vanilla `3d-force-graph` library is framework-agnostic and is the best fit.

#### Why 3d-force-graph

| Feature | Benefit |
|---|---|
| `emitParticle(link)` | Purpose-built API for one-shot message particle animation |
| `nodeThreeObject(node)` | Full Three.js `Object3D` per node — custom shapes, colors, animations |
| `linkDirectionalParticles` | Continuous particle flow for active channels |
| Dynamic `graphData()` | Read-mutate-write cycle for real-time add/remove |
| Scene access (`scene()`, `camera()`, `renderer()`) | Inject custom objects, post-processing (bloom), labels |
| Proven at scale | Comfortable at 100-200 nodes; examples with 4,000+ |

#### Alternatives Considered

| Library | Verdict |
|---|---|
| `react-force-graph-3d` | React wrapper — wrong framework for this codebase |
| `ngraph` | No built-in particle system; much more integration work |
| Raw Three.js | 10-20x dev effort to replicate what 3d-force-graph provides |
| 2D alternatives (d3, cytoscape) | Less visual impact; 3D better conveys the "living system" feel |

### Architecture

```
extras/agent-viz/
├── index.html              # Standalone entry point
├── package.json            # Vite + three + 3d-force-graph
├── tsconfig.json
├── vite.config.ts
├── src/
│   ├── main.ts             # Init, connect to hub, bootstrap graph
│   ├── graph.ts            # 3d-force-graph setup and update logic
│   ├── nodes.ts            # Node rendering (state → Three.js objects)
│   ├── messages.ts         # Message particle emission
│   ├── events.ts           # SSE client, event parsing, dispatch
│   ├── api.ts              # REST API calls (initial agent list, details)
│   ├── replay.ts           # Log file replay engine (stretch)
│   ├── ui.ts               # HUD overlays (agent detail panel, controls)
│   └── types.ts            # TypeScript interfaces
├── public/
│   └── textures/           # Glow sprites, icons
└── README.md
```

### Standalone vs Embedded

The visualizer should be a **standalone SPA** in `extras/agent-viz/` that:

1. Accepts hub URL and grove ID via URL params or config
2. Authenticates using the same token/session mechanism as the web frontend
3. Can optionally be iframe-embedded in the main web UI later

This avoids coupling to the Lit-based web app while keeping integration possible.

### Data Flow

```
Hub SSE (/events?sub=grove.{id}.>)
    │
    ▼
EventSource client (src/events.ts)
    │
    ├─► Agent created  → graph.addNode(agent)
    ├─► Agent deleted  → graph.removeNode(agentId)
    ├─► Agent status   → graph.updateNode(agentId, state)
    ├─► Message event  → graph.emitParticle(senderNode, recipientNode, msgType)
    └─► Tool call      → graph.showFileIndicator(agentId, toolName, fileName)
    │
    ▼
3d-force-graph instance
    │
    ▼
Three.js WebGL canvas
```

## Open Questions

### Data & API

1. **Message events via SSE:** The current SSE event subjects don't include a dedicated message-sent event. Messages are logged to `scion-messages` but not published via the `EventPublisher`. **Should we add a `PublishMessageSent` event to the hub's event system?** Without this, the visualizer can only show messages in replay mode from logs.

2. **File activity events:** Tool calls (file reads/writes) are logged in `scion-agents` but not published via SSE. **Should we add tool-call events to the SSE stream**, or is this too noisy? An alternative is polling the agent's recent log entries.

3. **Agent-to-agent link discovery:** The graph needs edges between agents that communicate. Should links be:
   - Created dynamically when a message is first observed between two agents?
   - Pre-populated from some declared topology (e.g., orchestrator → workers)?
   - Both (declared links + dynamic discovery)?

### Visualization

4. **Node layout strategy:** Force-directed is the default, but for orchestrator-worker patterns a hierarchical layout may be clearer. Should we support both and let users toggle? Or auto-detect based on messaging patterns?

5. **Scale target:** What's the realistic upper bound for agents in a single grove? 5-10 (typical)? 50? 200? This affects whether we need LOD (level-of-detail) optimizations.

6. **2D fallback:** Should the tool also offer a simpler 2D mode (using `force-graph-2d`) for lower-end devices or accessibility? The library supports this with the same API.

### UX & Integration

7. **Authentication:** The visualizer needs hub access. Should it:
   - Reuse the web app's session cookie (requires same-origin or CORS)?
   - Accept a bearer token via URL fragment?
   - Support an unauthenticated "demo/replay" mode from log files?

8. **Embedding:** Should the visualizer be linkable from the grove detail page in the web UI? If so, as an iframe or a web component?

9. **Replay mode priority:** How important is the ability to replay from exported JSON logs (like the `.scratch/downloaded-logs-*.json` files) vs. live-only visualization? Replay would be valuable for demos, debugging, and post-mortems.

10. **Audio/haptics:** Should message arrivals or state changes produce subtle audio cues? This could enhance the "living system" feel but may be distracting.

### Deployment

11. **Build and serve:** Should the visualizer be:
    - A fully static build (deploy to any CDN/file server)?
    - Served by the hub's web server alongside the main UI?
    - A dev-only tool run locally with `npm run dev`?

12. **Configuration:** Beyond hub URL and grove ID, what should be configurable? Candidate settings:
    - Color scheme / theme
    - Physics parameters (force strength, damping)
    - Which event types to visualize
    - Node label content (name vs slug vs template)

## MVP Scope

A minimal viable version would include:

1. Connect to hub SSE, seed graph from `GET /agents`
2. Add/remove nodes on agent create/delete events
3. Color nodes by phase (running=green, error=red, stopped=gray)
4. Animate activity with simple glow/pulse
5. Replay mode from exported log JSON files (since message SSE events don't exist yet)
6. Click-to-inspect agent detail overlay

**Deferred to v2:**
- Live message particle visualization (requires hub event system changes)
- File activity indicators
- Hierarchical layout mode
- Embedding in main web UI
- Audio cues
