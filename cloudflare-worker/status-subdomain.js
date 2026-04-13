/**
 * Pingr — Status page subdomain router
 *
 * Routes *.getpingr.com → getpingr.com/status/:username
 *
 * Example:
 *   acme-corp.getpingr.com  →  getpingr.com/status/acme-corp
 *   my-app.getpingr.com     →  getpingr.com/status/my-app
 *
 * Deploy:
 *   1. Create a Worker in Cloudflare dashboard → Workers & Pages → Create
 *   2. Paste this file as the worker code
 *   3. Add a route: *.getpingr.com/* → this worker
 *      (Settings → Triggers → Add route)
 */

const ROOT_DOMAIN   = 'getpingr.com'
const APEX_DOMAINS  = new Set([ROOT_DOMAIN, `www.${ROOT_DOMAIN}`])

export default {
  async fetch(request) {
    const url  = new URL(request.url)
    const host = url.hostname.toLowerCase()

    // If this is the root domain or www — pass through to Pages as normal
    if (APEX_DOMAINS.has(host)) {
      return fetch(request)
    }

    // Extract the subdomain: "acme-corp.getpingr.com" → "acme-corp"
    const subdomain = host.replace(`.${ROOT_DOMAIN}`, '')

    // Ignore non-status subdomains (api, www, etc.)
    const RESERVED = new Set(['api', 'www', 'mail', 'smtp'])
    if (RESERVED.has(subdomain) || !subdomain) {
      return fetch(request)
    }

    // Build the target URL: getpingr.com/status/acme-corp
    // Preserve any path/query beyond the root (e.g. for future sub-pages)
    const targetPath = url.pathname === '/' ? '' : url.pathname
    const targetURL  = `https://${ROOT_DOMAIN}/status/${subdomain}${targetPath}${url.search}`

    // Fetch from the main domain and return as-is
    const response = await fetch(targetURL, {
      headers: request.headers,
      method:  request.method,
    })

    // Clone the response so we can modify headers
    const newResponse = new Response(response.body, response)

    // Tell the browser this is the canonical URL for the subdomain
    newResponse.headers.set('X-Pingr-Subdomain', subdomain)

    return newResponse
  },
}
