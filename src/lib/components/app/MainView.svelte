<script lang="ts">
  interface Team {
    name: string;
    subnetId: number | null;
    targets?: string[];
  }

  interface Props {
    teams?: Team[];
  }

  let { teams = [] }: Props = $props();
</script>

<section class="rounded-md border border-white/10 bg-[#0d1316] p-5 text-white">
  <p class="text-xs font-semibold uppercase tracking-[0.22em] text-teal-200/60">Main</p>
  <h2 class="mt-2 text-xl font-semibold">Team roster</h2>

  <div class="mt-5 grid gap-3">
    {#each teams as team (team.name)}
      <article class="rounded-md border border-white/10 bg-white/[0.04] p-4">
        <h3 class="font-semibold">{team.name}</h3>
        <p class="mt-1 text-sm text-white/55">Subnet {team.subnetId ?? "unassigned"}</p>
      </article>
    {/each}
  </div>
</section>
