---
title: Agent Development Kit (ADK)
description: Developing custom autonomous agents with the Scion ADK.
---

Scion provides an Agent Development Kit (ADK) to facilitate the creation and integration of custom autonomous agents within the Scion ecosystem.

## Overview

The ADK streamlines the process of building agents that don't rely on the standard pre-packaged harnesses (like Gemini CLI or Claude Code), but instead use custom logic or alternative frameworks, while still benefiting from Scion's orchestration, workspace management, and observability features.

## ADK Runner Entrypoint

Scion includes a specialized runner entrypoint designed specifically for ADK agents. This entrypoint provides native support for the `--input` flag, facilitating more robust automated execution and easier testing of agent behaviors.

## Example Project

To get started quickly, Scion provides a complete example and Docker template located in the `examples/adk_scion_agent/` directory of the repository.

This example demonstrates:
- How to structure a custom agent.
- How to build a compatible Docker image.
- How to define the corresponding Scion template (`scion-agent.yaml`).
- How to handle inputs and interact with the Scion environment.

## Integration Points

When building an ADK agent, you will primarily interact with Scion through:
1.  **Environment Variables**: Scion injects configuration and context via environment variables (e.g., `SCION_AGENT_ID`, `SCION_WORKSPACE`).
2.  **Workspace Mount**: The designated workspace directory where your agent should perform its file operations.
3.  **Standard IO/sciontool**: Using the `sciontool` utility (injected into the container) to report status (`sciontool status`) and log structured messages.
