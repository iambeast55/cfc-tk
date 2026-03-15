import type { PageServerLoad, Actions } from './$types';
import { fail } from '@sveltejs/kit';

const BACKEND_URL = 'http://localhost:8080'; // Replace with your backend URL

export const load: PageServerLoad = async () => {
	try {
		const response = await fetch(`${BACKEND_URL}/api/teams`);
		if (!response.ok) throw new Error('Failed to fetch teams');
		const teams = await response.json();
		console.log("Teams from backend:", teams);

		return { teams };
	} catch (error) {
		console.error('Error loading teams:', error);
		return { teams: [] };
	}
};

export const actions: Actions = {
	addTeam: async ({ request }) => {
			const formData = await request.formData();
			// support multiple possible input names (teamName, name, Name)
			const teamName =
				(formData.get('teamName') as string) ||
				(formData.get('name') as string) ||
				(formData.get('Name') as string) ||
				null;

			if (!teamName) {
				return { success: false, error: 'Missing team name' };
			}

			// backend expects subnetId as integer; parse and only include when valid
			let subnetId: number | undefined = undefined;
			if(formData.get('subnetId') != null ) {
				const subnetRaw = formData.get('subnetId') as string;
				const parsed = Number(subnetRaw);
				if (!Number.isNaN(parsed) && Number.isInteger(parsed)) {
					subnetId = parsed;
				} else {
					return { success: false, error: 'subnetId must be an integer' };
				}
			}

			try {
				const payload: Record<string, unknown> = { name: teamName };
				if (subnetId !== undefined) payload.subnetId = subnetId;
				console.log("parsedSubnetId:", subnetId, "isNaN:", Number.isNaN(subnetId));
				const response = await fetch(`${BACKEND_URL}/api/teams`, {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(payload)
				});

				// if backend responds with non-2xx, include body text/json (if present) for debugging
				if (!response.ok) {
					let details = '';
					try {
						const body = await response.text();
						details = body ? `: ${body}` : '';
					} catch (e) {
						details = '';
					}
					console.error(`Backend error adding team${details}`);
					return { success: false, error: `Failed to add team${details}` };
				}

				const newTeam = await response.json();

				// Refetch all teams to keep UI in sync
				const teamsResponse = await fetch(`${BACKEND_URL}/api/teams`);
				const teams = teamsResponse.ok ? await teamsResponse.json() : [];

				return { success: true, teams };
			} catch (error) {
				console.error('Error adding team:', error);
				return { success: false, error: (error as Error).message || 'Failed to add team' };
			}
	},

	updateTeam: async ({ request }) => {
    const formData = await request.formData();

    const name = (formData.get('name') as string) || '';
    const subnetRaw = (formData.get('subnetId') as string) || '';

    if (!name) {
      return fail(400, { success: false, error: 'Missing team name' });
    }

    let subnetId: number | null = null;
    if (subnetRaw.trim() !== '') {
      const parsed = Number(subnetRaw);
      if (!Number.isInteger(parsed)) {
        return fail(400, { success: false, error: 'subnetId must be an integer' });
      }
      subnetId = parsed;
    }

    const response = await fetch(
      `${BACKEND_URL}/api/teams/${encodeURIComponent(name)}`,
      {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, subnetId })
      }
    );

    if (!response.ok) {
      const details = await response.text().catch(() => '');
      return fail(response.status, { success: false, error: details || 'Failed to update team' });
    }

    return { success: true };
  },

  deleteTeam: async ({ request }) => {
    const formData = await request.formData();

    const name = (formData.get('name') as string) || '';

    if (!name) {
      return fail(400, { success: false, error: 'Missing team name' });
    }
    const response = await fetch(
      `${BACKEND_URL}/api/teams/${encodeURIComponent(name)}`,
      {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name})
      }
    );

    if (!response.ok) {
      const details = await response.text().catch(() => '');
      return fail(response.status, { success: false, error: details || 'Failed to delete team' });
    }
    return { success: true };
  }
};

