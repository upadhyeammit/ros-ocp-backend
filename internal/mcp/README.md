# ROS OCP MCP Server

This package implements an [MCP](https://modelcontextprotocol.io/) (Model Context Protocol) server that exposes the ROS OCP REST API as tools for AI agents. Data access is **user-scoped**: the server forwards the client's `X-Rh-Identity` (or `Authorization: Bearer <identity>`) to the REST API on every call, so users only see their own org's data—aligned with the REST API behavior.

## Running the MCP server

From the repo root:

```bash
./rosocp start mcp
```

Or with custom port and API URL:

```bash
MCP_PORT=8090 ROS_API_BASE_URL=http://localhost:8000 ./rosocp start mcp
```

- **MCP_PORT**: Port the MCP server listens on (default: 8090).
- **ROS_API_BASE_URL**: Base URL of the ROS REST API that the MCP server calls (default: http://localhost:8000). When running alongside the API server, use the same host.

## Authentication

Clients (e.g. tech-sales-agent) must send **X-Rh-Identity** (base64-encoded identity, same as the REST API) on the HTTP request when connecting to the MCP server. Alternatively, **Authorization: Bearer &lt;identity&gt;** is accepted and forwarded as the identity value. If neither is present, the session is rejected (400 Bad Request).

## Tools

| Tool | Description |
|------|-------------|
| **list_recommendations** | List resource optimization recommendations. Supports filters: cluster, workload_type, workload, container, project, start_date, end_date; limit/offset for pagination; order_by, order_how; cpu_unit, memory_unit, true_units. |
| **get_recommendation** | Get a single recommendation by UUID. Optional: cpu_unit, memory_unit, true_units. |

Responses are the same JSON as the REST API.

## Transport

The server uses the MCP **Streamable HTTP** transport. Agents (e.g. LangChain with `langchain-mcp-adapters`, `transport="streamable_http"`) can connect to `http://<host>:<MCP_PORT>/` to list and call tools.
