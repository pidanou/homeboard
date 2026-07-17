<script lang="ts">
	import { Time } from '@internationalized/date';
	import * as TimeField from '$lib/components/ui/time-field';
	import { Input } from '$lib/components/ui/input';
	import { cn } from '$lib/utils.js';
	import { currentUser } from '$lib/stores/user.svelte';
	import { getLocale } from '$lib/paraglide/runtime';

	let {
		value = $bindable(''),
		disabled = false,
		id,
		class: className,
	}: {
		value?: string;
		disabled?: boolean;
		id?: string;
		class?: string;
	} = $props();

	function parseTime(v: string): Time | undefined {
		const parts = v.split(':');
		if (parts.length !== 2) return undefined;
		const [h, m] = parts.map(Number);
		if (Number.isNaN(h) || Number.isNaN(m)) return undefined;
		return new Time(h, m);
	}

	function formatTime(t: Time | undefined): string {
		if (!t) return '';
		return `${String(t.hour).padStart(2, '0')}:${String(t.minute).padStart(2, '0')}`;
	}

	const hourCycle = $derived(
		currentUser.value?.time_format === '12' ? 12 : currentUser.value?.time_format === '24' ? 24 : undefined
	);

	let timeValue = $state<Time | undefined>(parseTime(value));

	$effect(() => {
		if (formatTime(timeValue) !== value) timeValue = parseTime(value);
	});

	function onTimeFieldChange(t: Time | undefined) {
		const formatted = formatTime(t);
		if (formatted === value) return;
		timeValue = t;
		value = formatted;
	}
</script>

<!-- Mobile: native picker is the better experience here (OS wheel/clock UI). -->
<Input type="time" bind:value {disabled} {id} class={cn('md:hidden', className)} />

<!-- Desktop: segmented field, avoids the browser's native time input chrome. -->
<div class="hidden md:block">
	<TimeField.Root
		value={timeValue}
		onValueChange={onTimeFieldChange}
		{hourCycle}
		locale={getLocale()}
		granularity="minute"
		{disabled}
	>
		<TimeField.Input {id} class={className} />
	</TimeField.Root>
</div>
