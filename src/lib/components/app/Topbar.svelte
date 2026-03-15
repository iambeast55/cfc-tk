<script lang="ts">
  
  import { Button } from "$lib/components/ui/button";
  import { Plus } from "lucide-svelte";
  import SunIcon from "@lucide/svelte/icons/sun";
  import MoonIcon from "@lucide/svelte/icons/moon";
  import * as Tooltip from "$lib/components/ui/tooltip/index.js";
  import { toggleMode } from "mode-watcher";
  export let newTeamMemberFlyOut: () => void
</script>

<header class="sticky top-0 z-10 h-14 border-b bg-background/80 backdrop-blur">
  <div class="flex h-full items-center gap-3 px-4">
    <!-- Left spacer (matches SB Admin look with sidebar) -->
    <div class="hidden w-64 md:block"></div>

    <!-- Search
    <div class="relative flex max-w-xl flex-1 items-center">
      <Search class="absolute left-3 h-4 w-4 text-muted-foreground" />
      <Input class="pl-9" placeholder="Search for..." />
      <Button class="ml-2 shrink-0" size="icon" aria-label="Search">
        <Search class="h-4 w-4" />
      </Button>
    </div> -->

    <!-- Right actions -->
    <div class="ml-auto flex items-center gap-2">
      <Tooltip.Provider>
        <Tooltip.Root>
          <Tooltip.Trigger>
            <Button onclick={newTeamMemberFlyOut} variant="ghost" size="icon" aria-label="NewTeam" ><Plus class="h-4 w-4" /><span class="sr-only">New Team</span></Button>
          </Tooltip.Trigger>
            <Tooltip.Content><p>New Team</p> </Tooltip.Content>
        </Tooltip.Root>
      </Tooltip.Provider>
      
        <Button onclick={toggleMode} variant="outline" size="icon">
        <SunIcon
            class="h-[1.2rem] w-[1.2rem] scale-100 rotate-0 !transition-all dark:scale-0 dark:-rotate-90"
        />
        <MoonIcon
            class="absolute h-[1.2rem] w-[1.2rem] scale-0 rotate-90 !transition-all dark:scale-100 dark:rotate-0"
        />
        <span class="sr-only">Toggle theme</span>
    </Button>
    </div>
  </div>
</header>
