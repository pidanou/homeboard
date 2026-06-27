<script lang="ts">
	import type { Member, AppCategory } from '$lib/types';
	import { calDateToISO, fmtCalDate, calDateTimeToISO, rangeLabelFor } from '$lib/dates';
	import { api } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Select from '$lib/components/ui/select';
	import * as Popover from '$lib/components/ui/popover';
	import { Calendar } from '$lib/components/ui/calendar';
	import { RangeCalendar } from '$lib/components/ui/range-calendar';
	import type { DateRange } from 'bits-ui';
	import { CalendarDate, type DateValue } from '@internationalized/date';
	import { CalendarDays } from 'lucide-svelte';
	import CategoryPicker from '$lib/components/CategoryPicker.svelte';
	import IconPicker from '$lib/components/IconPicker.svelte';

	let { familyID, members, categories, onCreated }: {
		familyID: string;
		members: Member[];
		categories: AppCategory[];
		onCreated: () => void;
	} = $props();

	let isOpen = $state(false);
	let createType = $state<'task' | 'event' | 'birthday'>('task');
	let cf = $state({
		title: '', description: '', important: false,
		allDay: false, location: '', assignedTo: '', attendeeIDs: [] as string[],
	});
	let cfDueDate = $state<CalendarDate | undefined>(undefined);
	let cfDueOpen = $state(false);
	let cfDueTime = $state('');
	let cfEventRange = $state<DateRange>({ start: undefined, end: undefined });
	let cfStartTime = $state('09:00');
	let cfEndTime = $state('10:00');
	let cfEventPickerOpen = $state(false);
	let cfCategoryID = $state<string | undefined>(undefined);
	let cfBirthdayOf = $state('');
	let cfShowMore = $state(false);

	const REPEAT_LABELS: Record<string, string> = {
		none: 'Does not repeat', daily: 'Daily', weekly: 'Weekly', monthly: 'Monthly', yearly: 'Yearly'
	};
	let cfRepeat = $state<'none' | 'daily' | 'weekly' | 'monthly' | 'yearly'>('none');
	let cfIcon = $state<string | undefined>(undefined);

	const RRULE: Record<string, string> = {
		daily: 'FREQ=DAILY',
		weekly: 'FREQ=WEEKLY',
		monthly: 'FREQ=MONTHLY',
		yearly: 'FREQ=YEARLY',
	};

	function toCalDate(d: Date) {
		return new CalendarDate(d.getFullYear(), d.getMonth() + 1, d.getDate());
	}
	function formatTime(d: Date) {
		return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
	}

	export function open(t: 'task' | 'event' = 'task', start?: Date, end?: Date, allDay = false) {
		createType = t;
		cf = { title: '', description: '', important: false, allDay, location: '', assignedTo: '', attendeeIDs: [] };
		const cd = start ? toCalDate(start) : undefined;
		cfDueDate = cd;
		cfEventRange = cd ? { start: cd, end: end ? toCalDate(end) : cd } : { start: undefined, end: undefined };
		cfStartTime = start && !allDay ? formatTime(start) : '09:00';
		cfEndTime = end && !allDay ? formatTime(end) : '10:00';
		cfCategoryID = undefined;
		cfBirthdayOf = '';
		cfRepeat = 'none';
		cfIcon = undefined;
		cfShowMore = false;
		isOpen = true;
	}

	function toggleAttendee(ids: string[], uid: string): string[] {
		return ids.includes(uid) ? ids.filter((id) => id !== uid) : [...ids, uid];
	}

	async function submit() {
		const isBirthday = createType === 'birthday';
		if (isBirthday) cf.title = cfBirthdayOf.trim() + "'s Birthday";
		if (!cf.title.trim()) return;
		try {
			if (createType === 'task') {
				await api.post(`/api/v1/households/${familyID}/tasks`, {
					title: cf.title.trim(),
					description: cf.description,
					important: cf.important,
					assigned_to: cf.assignedTo || undefined,
					end_date: cfDueDate ? (cfDueTime ? calDateTimeToISO(cfDueDate, cfDueTime, false) : calDateToISO(cfDueDate)) : undefined,
					category_id: cfCategoryID,
					icon: cfIcon,
				});
			} else {
				if (!isBirthday && !cfEventRange.start) return;
				const isAllDay = cf.allDay || isBirthday;
				const cfEnd = cfEventRange.end ?? cfEventRange.start;
				const startCal = isBirthday ? cfDueDate! : cfEventRange.start!;
				const endCal = isBirthday ? cfDueDate! : (cfEnd ?? cfEventRange.start!);
				// All-day end is exclusive (iCal convention) — add 1 day so a single-day event has duration
				const savedEnd = isAllDay ? endCal.add({ days: 1 }) : endCal;
				await api.post(`/api/v1/households/${familyID}/events`, {
					title: cf.title.trim(),
					description: cf.description,
					location: cf.location,
					start_at: calDateTimeToISO(startCal, cfStartTime, isAllDay),
					end_at: calDateTimeToISO(savedEnd, cfEndTime, isAllDay),
					all_day: isAllDay,
					attendee_ids: cf.attendeeIDs,
					category_id: cfCategoryID,
					recurrence_rule: isBirthday ? RRULE['yearly'] : (cfRepeat !== 'none' ? RRULE[cfRepeat] : undefined),
					icon: cfIcon,
					important: cf.important,
					birthday_of: isBirthday ? cfBirthdayOf.trim() : undefined,
				});
			}
			isOpen = false;
			onCreated();
		} catch { }
	}
</script>

<Dialog.Root bind:open={isOpen}>
	<Dialog.Portal>
		<Dialog.Overlay />
		<Dialog.Content class="sm:max-w-md flex flex-col max-h-[90dvh]">
			<Dialog.Header>
				<Dialog.Title>New ticket</Dialog.Title>
			</Dialog.Header>

			<div class="flex flex-col gap-3 py-2 overflow-y-auto flex-1 min-h-0 px-1"
				onkeydown={(e) => { if (e.key === 'Enter' && (e.target as HTMLElement).tagName !== 'TEXTAREA') { e.preventDefault(); submit(); } }}>
				<!-- Type switcher: compact pills -->
				<div class="flex gap-1.5">
					<button
						class="px-3 py-1 rounded-full text-sm font-medium transition-colors cursor-pointer
							{createType === 'task' ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:text-foreground'}"
						onclick={() => (createType = 'task')}
					>Task</button>
					<button
						class="px-3 py-1 rounded-full text-sm font-medium transition-colors cursor-pointer
							{createType === 'event' ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:text-foreground'}"
						onclick={() => { createType = 'event'; cf.allDay = false; }}
					>Event</button>
				<button
					class="px-3 py-1 rounded-full text-sm font-medium transition-colors cursor-pointer
						{createType === 'birthday' ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:text-foreground'}"
					onclick={() => { createType = 'birthday'; cfBirthdayOf = ''; cfDueDate = undefined; }}
					>Birthday</button>
				</div>

				<!-- Title -->
				{#if createType === 'birthday'}
					<Input bind:value={cfBirthdayOf} placeholder="Person's name…" class="flex-1" />
				{:else}
					<div class="flex gap-2">
						<IconPicker bind:value={cfIcon} />
						<Input bind:value={cf.title} placeholder={createType === 'task' ? 'Buy groceries…' : 'Team dinner…'} class="flex-1" />
					</div>
				{/if}

				{#if createType === 'task'}
					<!-- Task primary: important + due date -->
					<div class="flex items-center gap-2">
						<label class="flex items-center gap-2 text-sm cursor-pointer shrink-0">
							<Checkbox bind:checked={cf.important} />
							Important
						</label>
						<Popover.Root bind:open={cfDueOpen}>
							<Popover.Trigger class="flex-1">
								<Button variant="outline" class="w-full justify-start gap-2 font-normal text-sm">
									<CalendarDays class="w-4 h-4 text-muted-foreground shrink-0" />
									{cfDueDate ? fmtCalDate(cfDueDate) : 'No due date'}
								</Button>
							</Popover.Trigger>
							<Popover.Content class="w-auto p-0" align="start">
								<Calendar type="single" bind:value={cfDueDate} onValueChange={() => (cfDueOpen = false)} />
							</Popover.Content>
						</Popover.Root>
						{#if cfDueDate}
							<Input type="time" bind:value={cfDueTime} class="w-32 shrink-0" />
						{/if}
					</div>

					<!-- Task secondary -->
					{#if cfShowMore}
						{#if members.length > 0}
							<Select.Root type="single" bind:value={cf.assignedTo}>
								<Select.Trigger class="w-full">{members.find(m => m.user_id === cf.assignedTo)?.name ?? 'Unassigned'}</Select.Trigger>
								<Select.Content>
									<Select.Item value="">Unassigned</Select.Item>
									{#each members as m}
										<Select.Item value={m.user_id}>{m.name}</Select.Item>
									{/each}
								</Select.Content>
							</Select.Root>
						{/if}
						<Textarea bind:value={cf.description} placeholder="Notes…" rows={2} />
						<CategoryPicker {familyID} {categories} bind:selectedID={cfCategoryID} />
					{/if}
				{:else if createType === 'birthday'}
					<!-- Birthday: single date picker -->
					<Popover.Root bind:open={cfDueOpen}>
						<Popover.Trigger class="flex-1">
							<Button variant="outline" class="w-full justify-start gap-2 font-normal text-sm">
								<CalendarDays class="w-4 h-4 text-muted-foreground shrink-0" />
								{cfDueDate ? fmtCalDate(cfDueDate) : 'Birthday date…'}
							</Button>
						</Popover.Trigger>
						<Popover.Content class="w-auto p-0" align="start">
							<Calendar type="single" bind:value={cfDueDate} onValueChange={() => (cfDueOpen = false)} />
						</Popover.Content>
					</Popover.Root>
				{:else}
					<!-- Event primary: dates, all day, times -->
					<Popover.Root bind:open={cfEventPickerOpen}>
						<Popover.Trigger>
							<Button variant="outline" class="w-full justify-start gap-2 font-normal text-sm">
								<CalendarDays class="w-4 h-4 text-muted-foreground shrink-0" />
								{rangeLabelFor(cfEventRange)}
							</Button>
						</Popover.Trigger>
						<Popover.Content class="w-auto p-0" align="start">
							<RangeCalendar
								bind:value={cfEventRange}
								onValueChange={() => { if (cfEventRange.start && cfEventRange.end) cfEventPickerOpen = false; }}
							/>
						</Popover.Content>
					</Popover.Root>
					<div class="flex items-center gap-4">
						<label class="flex items-center gap-2 text-sm cursor-pointer">
							<Checkbox bind:checked={cf.allDay} />
							All day
						</label>
						<label class="flex items-center gap-2 text-sm cursor-pointer">
							<Checkbox bind:checked={cf.important} />
							Important
						</label>
					</div>
					{#if !cf.allDay}
						<div class="flex gap-2">
							<Input type="time" bind:value={cfStartTime} class="flex-1" />
							<Input type="time" bind:value={cfEndTime} class="flex-1" />
						</div>
					{/if}

					<!-- Event secondary -->
					{#if cfShowMore}
						<Select.Root type="single" bind:value={cfRepeat}>
							<Select.Trigger class="w-full">{REPEAT_LABELS[cfRepeat] ?? 'Does not repeat'}</Select.Trigger>
							<Select.Content>
								<Select.Item value="none">Does not repeat</Select.Item>
								<Select.Item value="daily">Daily</Select.Item>
								<Select.Item value="weekly">Weekly</Select.Item>
								<Select.Item value="monthly">Monthly</Select.Item>
								<Select.Item value="yearly">Yearly</Select.Item>
							</Select.Content>
						</Select.Root>
						<Input bind:value={cf.location} placeholder="Location…" />
						{#if members.length > 0}
							<div class="flex flex-col gap-1.5">
								{#each members as m}
									<label class="flex items-center gap-2 text-sm cursor-pointer">
										<Checkbox
											checked={cf.attendeeIDs.includes(m.user_id)}
											onCheckedChange={() => (cf.attendeeIDs = toggleAttendee(cf.attendeeIDs, m.user_id))}
										/>
										{m.name}
									</label>
								{/each}
							</div>
						{/if}
						<Textarea bind:value={cf.description} placeholder="Notes…" rows={2} />
						<CategoryPicker {familyID} {categories} bind:selectedID={cfCategoryID} />
					{/if}
				{/if}

				<button
					class="text-xs text-muted-foreground hover:text-foreground transition-colors text-left w-fit"
					onclick={() => (cfShowMore = !cfShowMore)}
				>{cfShowMore ? '− Less options' : '+ More options'}</button>
			</div>

			<Dialog.Footer class="gap-2">
				<Button variant="outline" onclick={() => (isOpen = false)}>Cancel</Button>
				<Button onclick={submit} disabled={
						(createType === 'birthday' && (!cfBirthdayOf.trim() || !cfDueDate)) ||
						(createType === 'event' && (!cf.title.trim() || !cfEventRange.start)) ||
						(createType === 'task' && !cf.title.trim())
					}>
					Create
				</Button>
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
