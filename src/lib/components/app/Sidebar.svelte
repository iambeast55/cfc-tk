<script lang="ts">
  import { Button,buttonVariants } from "$lib/components/ui/button";
  import { Separator } from "$lib/components/ui/separator";
  import { APP_NAME } from "$lib/constants/ui";
  import logo from "$lib/assets/logo.png";
  import * as DropdownMenu from "$lib/components/ui/dropdown-menu/index.js"; 
  import * as Dialog from "$lib/components/ui/dialog/index.js"; 
  import { Label } from "$lib/components/ui/label/index.js";
  import { Input } from "$lib/components/ui/input/index.js";
  // Optional icons
  import {
    LayoutDashboard,
    NotebookPen,
    RectangleEllipsis
  } from "lucide-svelte";

  
  // simple “active” demo
  const active = "Dashboard";
  interface Team {
    name: string;
    subnetId: number | null;
  }

  interface Props {
    teams?: Team[];
    //modifyTeamMemberFlyOut: (modifyTeam: Team) => void;
  }

  let { teams = []}: Props = $props();
  let changesubnetId = $state(false)
  let deleteTeam = $state(false)
  let addTarget = $state(false)
  let editsubnetId = $state<number | null>(null);

  
</script>

<aside class="hidden w-64 shrink-0 border-r bg-background md:flex md:flex-col">
  <div class="flex h-14 items-center gap-3 border-b px-4">
  <img
    src={logo}
    alt="Red Team Village"
    class="h-16 w-auto"
  />
  <span class="text-sm font-semibold">{APP_NAME}</span>
</div>

  <nav class="flex-1 overflow-y-auto px-2 py-4 text-sm">
    <div class="px-2 pb-2 text-xs font-semibold text-muted-foreground">CORE</div>

    <Button
      variant={active === "Dashboard" ? "secondary" : "ghost"}
      class="w-full justify-start gap-2"
    >
      <LayoutDashboard class="h-4 w-4" />
      Dashboard
    </Button>

    <Separator class="my-4" />

    <div class="px-2 pb-2 text-xs font-semibold text-muted-foreground">TEAMS
    </div>
    {#each teams as modifyTeam (modifyTeam.name)}
    <DropdownMenu.Root>
      <DropdownMenu.Trigger>
        {#snippet child({ props })}
      <Button  {...props} variant="ghost" class="w-full justify-start gap-2">
        {modifyTeam.name}</Button>
      {/snippet}
      </DropdownMenu.Trigger>
      <DropdownMenu.Content class="w-52">
        <DropdownMenu.Item onSelect={()=> (changesubnetId = true)}>Change Subnet ID</DropdownMenu.Item>
        <DropdownMenu.Item onSelect={()=> (addTarget = true)}>New Target</DropdownMenu.Item>
        <DropdownMenu.Item>Delete Target</DropdownMenu.Item>
        <DropdownMenu.Item onSelect={()=> (deleteTeam = true)}>Delete Team</DropdownMenu.Item>
      </DropdownMenu.Content>
    </DropdownMenu.Root>

    <Dialog.Root bind:open={changesubnetId}>
        <Dialog.Content>
          <Dialog.Header>
            <Dialog.Title>Change Subnet ID for {modifyTeam.name}</Dialog.Title>
            <Dialog.Description>
              Update the Subnet ID for the selected team.
            </Dialog.Description>
          </Dialog.Header>
          <form method="POST" action="?/updateTeam">
            <div class="grid flex-1 auto-rows-min gap-6 px-4">
              <div class="grid gap-3">
                <Label for="subnetId">Subnet ID</Label>
                <input type="hidden" name="name" value={modifyTeam?.name ?? ''} />
                <Input id="subnetId" name="subnetId" bind:value={editsubnetId} placeholder={String(modifyTeam?.subnetId)}  required />
              </div>
              <Separator class="my-5 mb-2" />
            </div>
            <Dialog.Footer>
              <Button type="submit">Save changes <span class="sr-only">Save Changes</span></Button>  
              <Dialog.Close class={buttonVariants({ variant: "outline" })}>Cancel</Dialog.Close>
            </Dialog.Footer>
          </form>
        </Dialog.Content>
    </Dialog.Root>

    <Dialog.Root bind:open={deleteTeam}>
        <Dialog.Content>
          <Dialog.Header>
            <Dialog.Title>Delete Team: {modifyTeam.name}</Dialog.Title>
            <Dialog.Description>
              Are you sure you want to delete this team? This action cannot be undone.
            </Dialog.Description>
          </Dialog.Header>
          <form method="POST" action="?/deleteTeam">
            <div class="grid flex-1 auto-rows-min gap-6 px-4">
              <input type="hidden" name="name" value={modifyTeam?.name ?? ''} />
            </div>
            <Dialog.Footer>
              <Button type="submit">Delete Team <span class="sr-only">Delete Team</span></Button>  
              <Dialog.Close class={buttonVariants({ variant: "outline" })}>Cancel</Dialog.Close>
            </Dialog.Footer>
          </form>
        </Dialog.Content>
    </Dialog.Root>

    <Dialog.Root bind:open={addTarget}>
        <Dialog.Content>
          <Dialog.Header>
            <Dialog.Title>New Target for {modifyTeam.name}</Dialog.Title>
            <Dialog.Description>
              Add a new target for the selected team.
            </Dialog.Description>
          </Dialog.Header>
          <form method="POST" action="?/updateTeam">
           
          </form>
        </Dialog.Content>
    </Dialog.Root>
    {/each}

    <Separator class="my-4" />

    <div class="px-2 pb-2 text-xs font-semibold text-muted-foreground">EXTRAS</div>

    <Button variant="ghost" class="w-full justify-start gap-2">
      <NotebookPen  class="h-4 w-4" />
      Notes
    </Button>

    <Button variant="ghost" class="w-full justify-start gap-2">
      <RectangleEllipsis class="h-4 w-4" />
      Credentials
    </Button>
  </nav>

  <!-- Footer / Logged in -->
  <div class="border-t p-4">
    <div class="text-xs text-muted-foreground">Logged in as:</div>
    <div class="text-sm font-medium">SB Admin Svelte</div>
  </div>
</aside>
