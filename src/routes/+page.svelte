<script lang="ts">
  import Sidebar from "$lib/components/app/Sidebar.svelte";
  import Topbar from "$lib/components/app/Topbar.svelte";
  import * as Sheet from "$lib/components/ui/sheet/index.js";
  import { Label } from "$lib/components/ui/label/index.js";
  import { Input } from "$lib/components/ui/input";
  import { Button } from "$lib/components/ui/button";
  import {buttonVariants} from "$lib/components/ui/button/button.svelte"
  import MainView from "$lib/components/app/MainView.svelte";
  import { invalidateAll } from "$app/navigation";
  import { Separator } from "$lib/components/ui/separator/index.js";
  import * as Dialog from "$lib/components/ui/dialog/index.js";
  interface Team {
    name: string;
    subnetId: number | null;
  }


  let { data } = $props();
  let newTeamMemberFlyOutOpen = $state(false);
  let newTeam = $state<Team>({ name: "", subnetId: null });
  //let modifyTeamMemberFlyOutOpen = $state(false);
  //let modifyTeam = $state<Team | null>(null);
  //let editName = $state<string>("");
  //let editsubnetId = $state<string>("");

  /**function openModifyTeam(team: Team) {
  console.log("clicked team:", team, Object.keys(team));

    modifyTeam = team;
    editName = team.name ?? "";
    editsubnetId = team.subnetId != null ? String(team.subnetId) : "test";
    modifyTeamMemberFlyOutOpen = true;
  }**/

  const handleAddTeam = async (e: SubmitEvent) => {
    e.preventDefault();
    const form = e.target as HTMLFormElement;
    const formData = new FormData(form);
    
    try {
      const response = await fetch("?/addTeam", {
        method: "POST",
        body: formData
      });
      
      if (response.ok) {
        newTeamMemberFlyOutOpen = false;
        newTeam = { name: "", subnetId: null };
        await invalidateAll(); // reloads page data (load), so teams list refreshes
        
      } else {
        console.error("Failed to add team:", await response.text());
      }
        // Data will be revalidated automatically by SvelteKit's form actions, but you can also manually trigger it if needed
    } catch (error) {
      console.error("Error adding team:", error);
    }
  };
  
</script>

<div class="min-h-screen bg-muted/30">
  <div class="flex min-h-screen">
    <Sidebar teams={data.teams} />

    <div class="flex min-w-0 flex-1 flex-col">
      <Topbar newTeamMemberFlyOut={() => (newTeamMemberFlyOutOpen = true)} />
      <main class="min-h-0 flex-1 p-6">
        <MainView teams={data.teams} />
      </main>
    </div>
  </div>
</div>

<Sheet.Root bind:open={newTeamMemberFlyOutOpen}>
    <Sheet.Content side="right">
                  <Sheet.Header>
                    <Sheet.Title>Add a New Team</Sheet.Title>
                    <Sheet.Description>
                      Add each team you're responsible for.
                    </Sheet.Description>
                  </Sheet.Header>
                  <form onsubmit={handleAddTeam}>
                    <div class="grid flex-1 auto-rows-min gap-6 px-4">
                      <div class="grid gap-3">
                        <Label for="name">Name</Label>
                        <Input id="name" name="name" bind:value={newTeam.name} required />
                      </div>
                      <div class="grid gap-3">
                        <Label for="subnetId">Subnet Number</Label>
                        <Input id="subnetId" name="subnetId" bind:value={newTeam.subnetId} />
                      </div>
                      <Separator class="my-5 mb-2" />
                    </div>
                    <Sheet.Footer>
                      <Button type="submit">Save changes <span class="sr-only">Save Changes</span></Button>  
                      <Sheet.Close class={buttonVariants({ variant: "outline" })}>Cancel</Sheet.Close>
                    </Sheet.Footer>
                  </form>
    </Sheet.Content>
</Sheet.Root>


<!--<Sheet.Root bind:open={modifyTeamMemberFlyOutOpen}>
    <Sheet.Content side="right" class="w-full sm:w-[480px] md:w-[640px] lg:w-[800px] sm:max-w-none">
                  <Sheet.Header>
                    <Sheet.Title>Modify Team: {modifyTeam?.name}</Sheet.Title>
                    <Sheet.Description>
                      Update the details of the selected team.
                    </Sheet.Description>
                  </Sheet.Header>
                  <form method="POST" action="?/updateTeam">
                    <div class="grid flex-1 auto-rows-min gap-6 px-4">
                      <div class="grid gap-3">
                        <Label for="subnetId">Subnet</Label>
                        <input type="hidden" name="originalsubnetId" value={modifyTeam?.subnetId ?? ''} />
                        <Input id="subnetId" name="subnetId" bind:value={editsubnetId} placeholder={editsubnetId}/>
                      </div>
                      <Separator class="my-5 mb-2" />
                    </div>
                    <Sheet.Footer>
                      <Button type="submit">Save changes <span class="sr-only">Save Changes</span></Button>  
                      <Sheet.Close class={buttonVariants({ variant: "outline" })}>Cancel</Sheet.Close>
                    </Sheet.Footer>
                  </form>
    </Sheet.Content>
</Sheet.Root> -->