<script lang="ts">
	import { goto } from '$app/navigation';
	import { base } from '$app/paths';
	import { api } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as msg from '$lib/paraglide/messages.js';

	let name = $state('');
	let loading = $state(false);

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		loading = true;
		try {
			const family = await api.post<{ id: string }>('/api/v1/households', { name });
			goto(`${base}/households/${family.id}`);
		} catch { } finally {
			loading = false;
		}
	}
</script>

<div class="px-4 md:px-6 pt-6 pb-8">
	<div class="max-w-sm mx-auto">
		<h2 class="text-xl font-semibold mb-6">{msg.household_new_heading()}</h2>
		<form onsubmit={submit} class="flex flex-col gap-4">
			<div class="flex flex-col gap-1.5">
				<Label for="name">{msg.household_name_label()}</Label>
				<Input id="name" bind:value={name} required placeholder={msg.household_name_placeholder()} />
			</div>
			<Button type="submit" disabled={loading}>
				{loading ? msg.household_creating() : msg.dialog_create()}
			</Button>
		</form>
	</div>
</div>
