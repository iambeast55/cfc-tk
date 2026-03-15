<script lang="ts">
    import * as Select from "$lib/components/ui/select/index.js";
    import { Separator } from "$lib/components/ui/separator/index.js";
    import { onMount } from "svelte";
    import { browser } from "$app/environment";
    
    interface Team {
      name: string;
      subnetId: string;
    }

    interface Props {
        teams?: Team[];
    }

    let { teams = [] }: Props = $props();
    let selectedTeam = $state<string>("");
    let selectedTeamName = $state<string>("Select a Team");
    $effect(() => {
      if (!selectedTeam) {
        selectedTeamName = "Select a Team";
        return;
      }

      const found = (teams || []).find((t) => t.name === selectedTeam);
      selectedTeamName = found ? found.name : selectedTeam;
    });
    // debug: expose open state to programmatically open the dropdown
    let selectOpen = $state<boolean>(false);
    let terminalElement: HTMLElement;
    let term: any;

 onMount(async () => {
    if (!browser) return;

    const { Terminal } = await import("xterm");
    const { FitAddon } = await import("xterm-addon-fit");
    term = new Terminal({
      cursorBlink: true,
      fontSize: 13,
      fontFamily:
        "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace",
    });

    const fit = new FitAddon();
    term.loadAddon(fit);

    term.open(terminalElement);
    fit.fit();
    term.focus();

    const ro = new ResizeObserver(() => fit.fit());
    ro.observe(terminalElement);

    term.writeln("Hello from xterm.js");
    term.write("$ ");

    // simple local echo
    term.onData((data: string) => {
      // handle Enter
      if (data === "\r") {
        term.write("\r\n$ ");
        return;
      }
      // handle Backspace
      if (data === "\u007f") {
        term.write("\b \b");
        return;
      }
      // echo everything else
      term.write(data);
    });
  });
</script>
<svelte:head>
    <link rel="stylesheet" href="https://raw.githubusercontent.com/xtermjs/xterm.js/refs/heads/master/css/xterm.css">
</svelte:head>

<div class="grid grid-cols-4 gap-6">
  <div class="rounded-lg border p-6">
    <div class="flex items-center justify-center">
        <Select.Root type="single" bind:open={selectOpen} bind:value={selectedTeam}>
        <Select.Trigger class="w-[280px]">{selectedTeamName}</Select.Trigger>
        <Select.Content>
        <Select.Group>
        {#each teams as team (team.name)}
          <Select.Item value={team.name}>{team.name}</Select.Item>
        {/each}
        {#if teams.length === 0}
            <Select.Item value="" disabled>No teams available</Select.Item>
        {/if}
        </Select.Group>
        </Select.Content>
        </Select.Root>
    </div>
<Separator class="my-5 mb-2" />
<h1 class="text-sm text-left">Targets</h1>
</div>

  <div class="col-span-2 md:col-span-2 row-span-2 rounded-lg border p-6">
    <div class="h-[420px] w-full" bind:this={terminalElement} role="application"  onclick={() => term?.focus()}></div>
    <Separator class="my-5 mb-2" />
    <h1 class="text-lg text-center">Quick Commands</h1>
  </div>


  <div class="rounded-lg border p-6">Card 3</div>
</div>