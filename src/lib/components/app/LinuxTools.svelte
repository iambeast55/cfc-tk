<script lang="ts">
  import { Button } from "$lib/components/ui/button";
  let { teams }: { teams: { name: string }[] } = $props();
  interface Target { id: number; hostname: string; ip: string; os: string; }
  let team = $state("");
  let targets = $state<Target[]>([]);
  let targetId = $state("");
  let username = $state("");
  let port = $state(22);
  let authMode = $state("password");
  let password = $state("");
  let keyPath = $state("");
  let action = $state("ssh");
  let localPath = $state("");
  let remotePath = $state("");
  let loading = $state(false);
  let busy = $state(false);
  let error = $state("");
  let message = $state("");
  const api = "http://localhost:8765";
  $effect(() => {
    if (!teams.some((entry) => entry.name === team)) team = teams[0]?.name ?? "";
  });
  $effect(() => {
    const selected = team;
    targets = [];
    targetId = "";
    error = "";
    message = "";
    if (!selected) return;
    const controller = new AbortController();
    loading = true;
    void (async () => {
      try {
        const response = await fetch(`${api}/api/teams/${encodeURIComponent(selected)}/targets`, { signal: controller.signal });
        if (!response.ok) throw new Error((await response.text()).trim());
        const loaded = (await response.json()) as Target[];
        if (controller.signal.aborted) return;
        targets = loaded.filter((target) => target.os === "linux");
        targetId = targets[0] ? String(targets[0].id) : "";
      } catch (cause) {
        if (!controller.signal.aborted) error = cause instanceof Error ? cause.message : "Could not load targets.";
      } finally {
        if (!controller.signal.aborted) loading = false;
      }
    })();
    return () => controller.abort();
  });
  async function launch(event: SubmitEvent) {
    event.preventDefault();
    busy = true;
    error = "";
    message = "";
    try {
      const response = await fetch(`${api}/api/teams/${encodeURIComponent(team)}/linux/launch`, {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ targetId: Number(targetId), username, port, authMode,
          password: authMode === "password" ? password : "", keyPath, action, localPath, remotePath })
      });
      if (!response.ok) throw new Error((await response.text()).trim());
      const result = await response.json();
      password = "";
      message = `Opened ${result.title} in ${result.terminal}. ${action === "ssh" ? "Continue in the terminal." : "Transfer progress and completion appear in the terminal."}`;
    } catch (cause) {
      error = cause instanceof Error ? cause.message : "Could not open terminal.";
    } finally { busy = false; }
  }
</script>

<section class="linux-tools">
  <div>
    <p class="eyebrow">LINUX</p>
    <h2>SSH &amp; File Transfer</h2>
    <p class="muted">Terminals open on the Linux machine running CFCBench.</p>
    <form onsubmit={launch}>
      <fieldset disabled={busy}>
        <label>Team<select bind:value={team} disabled={!teams.length}>{#each teams as entry}<option value={entry.name}>{entry.name}</option>{/each}</select></label>
        <label>Linux target<select bind:value={targetId} disabled={loading || !targets.length}>
          {#if !targets.length}<option value="">{loading ? "Loading..." : "No Linux targets"}</option>{/if}
          {#each targets as target}<option value={String(target.id)}>{target.hostname} / {target.ip}</option>{/each}
        </select></label>
        <div class="pair">
          <label>Username<input bind:value={username} required autocomplete="off" placeholder="operator" /></label>
          <label>SSH port<input type="number" bind:value={port} min="1" max="65535" required /></label>
        </div>
        <label>Authentication<select bind:value={authMode}><option value="password">Password</option><option value="key">Private key</option></select></label>
        {#if authMode === "password"}
          <label>Password<input type="password" bind:value={password} autocomplete="new-password" placeholder="Leave blank to enter in terminal" /></label>
          <p class="muted">A supplied password requires sshpass. First-time hosts may require connecting with a blank password field to confirm the host key.</p>
        {:else}
          <label>Private key path<input bind:value={keyPath} required placeholder="/home/operator/.ssh/id_ed25519" /></label>
          <p class="muted">Use a key file on the CFCBench machine. Enter its passphrase in the terminal when prompted.</p>
        {/if}
        <label>Action<select bind:value={action}><option value="ssh">Open SSH terminal</option><option value="upload">Upload file with SCP</option><option value="download">Download file with SCP</option></select></label>
        {#if action !== "ssh"}
          <label>{action === "upload" ? "Local source file" : "Local destination file"}<input bind:value={localPath} required placeholder="/home/operator/files/report.txt" /></label>
          <label>{action === "upload" ? "Remote destination path" : "Remote source file"}<input bind:value={remotePath} required placeholder="/home/operator/report.txt" /></label>
          <p class="muted">Local paths refer to the CFCBench machine. Transfers can replace files at the destination.</p>
        {/if}
        <Button type="submit" disabled={busy || loading || !targetId} class="bg-teal-200 text-black hover:bg-teal-100">{busy ? "Opening..." : action === "ssh" ? "Open SSH terminal" : "Open transfer terminal"}</Button>
      </fieldset>
    </form>
  </div>
  <div aria-live="polite">
    <h3>Session</h3>
    {#if error}<p role="alert" class="error">{error}</p>{/if}
    {#if message}<p class="success">{message}</p>{:else}<p class="muted">Choose a target and connection details to begin.</p>{/if}
  </div>
</section>

<style>
  .linux-tools { display:grid; grid-template-columns:minmax(0, 440px) minmax(0,1fr); gap:32px; color:#e6edf0; }
  .linux-tools > div { min-width:0; }
  h2 { font-size:22px; font-weight:600; margin:6px 0 12px; }
  h3 { font-size:18px; font-weight:600; margin-bottom:12px; }
  .eyebrow { color:#9bbda9; font-size:12px; }
  fieldset { display:grid; min-width:0; gap:16px; margin-top:24px; }
  label { display:grid; min-width:0; gap:8px; font-size:13px; color:#bdc5c7; }
  input, select { width:100%; min-width:0; border:1px solid #ffffff24; background:#0d1316; color:#e6edf0; padding:11px 12px; border-radius:6px; font-size:14px; }
  input:focus-visible, select:focus-visible { outline:2px solid #82bfb5; outline-offset:2px; }
  .pair { display:grid; grid-template-columns: minmax(0,1fr) minmax(0,1fr); gap:12px; }
  .muted { color:#a5afb3; font-size:13px; line-height:1.6; }
  .error, .success { overflow-wrap:anywhere; font-size:14px; line-height:1.6; }
  .error { color:#f0b4c2; } .success { color:#bcd79c; }
  @media (max-width:800px) { .linux-tools { grid-template-columns:minmax(0,1fr); } }
</style>
