import os, sys, asyncio
from aiohttp import web, ClientSession, ClientTimeout

UPSTREAM = os.environ["UPSTREAM_URL"]
PORT     = int(os.environ.get("PORT", "8080"))
MAINT    = os.environ.get("MAINTENANCE_MODE", "").lower() in ("1", "true", "yes")

async def handle(request: web.Request) -> web.Response:
    if MAINT:
        return web.Response(status=503, text="Service Unavailable (maintenance mode)")
        
    path_qs = request.rel_url.raw_path
    async with ClientSession(timeout=ClientTimeout(total=30)) as sess:
        upstream_resp = await sess.request(
            method=request.method,
            url=f"{UPSTREAM}{path_qs}",
            headers={k: v for k, v in request.headers.items() if k.lower() != "host"},
            data=await request.read()
        )
        body = await upstream_resp.read()
    return web.Response(body=body, status=upstream_resp.status,
                        headers=upstream_resp.headers)

app = web.Application()
app.router.add_route("*", "/{tail:.*}", handle)

if __name__ == "__main__":
    web.run_app(app, port=PORT, access_log=None)