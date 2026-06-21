<script lang="ts">
	import type { Member, AppCategory } from '$lib/types';
	import { calDateToISO, fmtCalDate, calDateTimeToISO, rangeLabelFor } from '$lib/dates';
	import { api } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Select from '$lib/components/ui/select';
	import * as Popover from '$lib/components/ui/popover';
	import { Calendar } from '$lib/components/ui/calendar';
	import { RangeCalendar } from '$lib/components/ui/range-calendar';
	import { Select as SelectPrimitive } from 'bits-ui';
	import type { DateRange } from 'bits-ui';
	import { CalendarDate, type DateValue } from '@internationalized/date';
	import { CheckSquare, CalendarDays, Repeat, Cake } from 'lucide-svelte';
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
	let cfEventRange = $state<DateRange>({ start: undefined, end: undefined });
	let cfStartTime = $state('09:00');
	let cfEndTime = $state('10:00');
	let cfEventPickerOpen = $state(false);
	let cfCategoryID = $state<string | undefined>(undefined);
	let cfRepeat = $state<'none' | 'daily' | 'weekly' | 'monthly' | 'yearly'>('none');
	let cfIcon = $state<string | undefined>(undefined);

	const RRULE: Record<string, string> = {
		daily: 'FREQ=DAILY',
		weekly: 'FREQ=WEEKLY',
		monthly: 'FREQ=MONTHLY',
		yearly: 'FREQ=YEARLY',
	};

	export function open(t: 'task' | 'event' | 'birthday' = 'task') {
		createType = t;
		cf = { title: '', description: '', important: false, allDay: false, location: '', assignedTo: '', attendeeIDs: [] };
		cfDueDate = undefined;
		cfEventRange = { start: undefined, end: undefined };
		cfStartTime = '09:00';
		cfEndTime = '10:00';
		cfCategoryID = undefined;
		cfRepeat = 'none';
		cfIcon = undefined;
		isOpen = true;
	}

	function toggleAttendee(ids: string[], uid: string): string[] {
		return ids.includes(uid) ? ids.filter((id) => id !== uid) : [...ids, uid];
	}

	async function submit() {
		if (!cf.title.trim()) return;
		try {
			if (createType === 'task') {
				await api.post(`/api/v1/families/${familyID}/tasks`, {
					title: cf.title.trim(),
					description: cf.description,
					important: cf.important,
					assigned_to: cf.assignedTo || undefined,
					end_date: cfDueDate ? calDateToISO(cfDueDate) : undefined,
					category_id: cfCategoryID,
					icon: cfIcon,
				});
			} else if (createType === 'birthday') {
				if (!cfDueDate) return;
				await api.post(`/api/v1/families/${familyID}/events`, {
					title: cf.title.trim(),
					description: cf.description,
					start_at: calDateTimeToISO(cfDueDate, '00:00', true),
					end_at: calDateTimeToISO(cfDueDate, '00:00', true),
					all_day: true,
					attendee_ids: [],
					category_id: cfCategoryID,
					recurrence_rule: RRULE['yearly'],
					type: 'birthday',
				});
			} else {
				if (!cfEventRange.start) return;
				const cfEnd = cfEventRange.end ?? cfEventRange.start;
				await api.post(`/api/v1/families/${familyID}/events`, {
					title: cf.title.trim(),
					description: cf.description,
					location: cf.location,
					start_at: calDateTimeToISO(cfEventRange.start, cfStartTime, cf.allDay),
					end_at: calDateTimeToISO(cfEnd, cfEndTime, cf.allDay),
					all_day: cf.allDay,
					attendee_ids: cf.attendeeIDs,
					category_id: cfCategoryID,
					recurrence_rule: cfRepeat !== 'none' ? RRULE[cfRepeat] : undefined,
					icon: cfIcon,
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

			<div class="flex flex-col gap-4 py-2 overflow-y-auto flex-1 min-h-0 px-1">
				<div class="flex gap-2">
					<button
						class="flex-1 flex flex-col items-center gap-1.5 rounded-lg border-2 py-3 text-sm font-medium transition-colors cursor-pointer
							{createType === 'task' ? 'border-primary bg-primary/5 text-primary' : 'border-border text-muted-foreground hover:border-muted-foreground'}"
						onclick={() => (createType = 'task')}
					>
						<CheckSquare class="w-5 h-5" />Task
					</button>
					<button
						class="flex-1 flex flex-col items-center gap-1.5 rounded-lg border-2 py-3 text-sm font-medium transition-colors cursor-pointer
							{createType === 'event' ? 'border-blue-500 bg-blue-500/5 text-blue-600 dark:text-blue-400' : 'border-border text-muted-foreground hover:border-muted-foreground'}"
						onclick={() => (createType = 'event')}
					>
						<CalendarDays class="w-5 h-5" />Event
					</button>
					<button
						class="flex-1 flex flex-col items-center gap-1.5 rounded-lg border-2 py-3 text-sm font-medium transition-colors cursor-pointer
							{createType === 'birthday' ? 'border-pink-500 bg-pink-500/5 text-pink-600 dark:text-pink-400' : 'border-border text-muted-foreground hover:border-muted-foreground'}"
						onclick={() => (createType = 'birthday')}
					>
						<Cake class="w-5 h-5" />Birthday
					</button>
				</div>

				<div class="flex flex-col gap-1.5">
					<Label for="cf-title">Title</Label>
					<div class="flex gap-2">
						<IconPicker bind:value={cfIcon} />
						<Input id="cf-title" bind:value={cf.title} placeholder={createType === 'task' ? 'Buy groceries…' : 'Team dinner…'} class="flex-1" />
					</div>
				</div>

				<div class="flex flex-col gap-1.5">
					<Label for="cf-desc">Description</Label>
					<Textarea id="cf-desc" bind:value={cf.description} placeholder="Optional details…" rows={2} />
				</div>

				{#if createType === 'birthday'}
					<div class="flex flex-col gap-1.5">
						<Label>Date of birth</Label>
						<Popover.Root bind:open={cfDueOpen}>
							<Popover.Trigger>
								<Button variant="outline" class="w-full justify-start gap-2 font-normal text-sm">
									<Cake class="w-4 h-4 text-muted-foreground shrink-0" />
									{cfDueDate ? fmtCalDate(cfDueDate) : 'Pick a date'}
								</Button>
							</Popover.Trigger>
							<Popover.Content class="w-auto p-0" align="start">
								<Calendar type="single" bind:value={cfDueDate} onValueChange={() => (cfDueOpen = false)} />
							</Popover.Content>
						</Popover.Root>
					</div>
				{:else if createType === 'task'}
					<div class="flex gap-3">
						<div class="flex flex-col gap-1.5 flex-1">
							<label class="flex items-center gap-2 text-sm cursor-pointer mt-5">
								<Checkbox bind:checked={cf.important} />
								Important
							</label>
						</div>
						<div class="flex flex-col gap-1.5 flex-1">
							<Label>Due date</Label>
							<Popover.Root bind:open={cfDueOpen}>
								<Popover.Trigger>
									<Button variant="outline" class="w-full justify-start gap-2 font-normal text-sm">
										<CalendarDays class="w-4 h-4 text-muted-foreground shrink-0" />
										{cfDueDate ? fmtCalDate(cfDueDate) : 'Pick a date'}
									</Button>
								</Popover.Trigger>
								<Popover.Content class="w-auto p-0" align="start">
									<Calendar type="single" bind:value={cfDueDate} onValueChange={() => (cfDueOpen = false)} />
								</Popover.Content>
							</Popover.Root>
						</div>
					</div>
					{#if members.length > 0}
						<div class="flex flex-col gap-1.5">
							<Label>Assign to</Label>
							<Select.Root type="single" bind:value={cf.assignedTo}>
								<Select.Trigger class="w-full"><SelectPrimitive.Value placeholder="Unassigned" /></Select.Trigger>
								<Select.Content>
									<Select.Item value="">Unassigned</Select.Item>
									{#each members as m}
										<Select.Item value={m.user_id}>{m.name}</Select.Item>
									{/each}
								</Select.Content>
							</Select.Root>
						</div>
					{/if}
				{:else}
					<div class="flex flex-col gap-1.5">
						<Label>Dates</Label>
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
					</div>
					{#if !cf.allDay}
						<div class="flex gap-3">
							<div class="flex flex-col gap-1.5 flex-1">
								<Label for="cf-start-time">Start time</Label>
								<Input id="cf-start-time" type="time" bind:value={cfStartTime} />
							</div>
							<div class="flex flex-col gap-1.5 flex-1">
								<Label for="cf-end-time">End time</Label>
								<Input id="cf-end-time" type="time" bind:value={cfEndTime} />
							</div>
						</div>
					{/if}
					<label class="flex items-center gap-2 text-sm cursor-pointer">
						<Checkbox bind:checked={cf.allDay} />
						All day
					</label>
					<div class="flex flex-col gap-1.5">
						<Label>Repeat</Label>
						<Select.Root type="single" bind:value={cfRepeat}>
							<Select.Trigger class="w-full">
								<div class="flex items-center gap-2">
									<Repeat class="w-4 h-4 text-muted-foreground shrink-0" />
									<SelectPrimitive.Value placeholder="Does not repeat" />
								</div>
							</Select.Trigger>
							<Select.Content>
								<Select.Item value="none">Does not repeat</Select.Item>
								<Select.Item value="daily">Daily</Select.Item>
								<Select.Item value="weekly">Weekly</Select.Item>
								<Select.Item value="monthly">Monthly</Select.Item>
								<Select.Item value="yearly">Yearly</Select.Item>
							</Select.Content>
						</Select.Root>
					</div>
					<div class="flex flex-col gap-1.5">
						<Label for="cf-location">Location</Label>
						<Input id="cf-location" bind:value={cf.location} placeholder="Optional location…" />
					</div>
					{#if members.length > 0}
						<div class="flex flex-col gap-1.5">
							<Label>Members</Label>
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
						</div>
					{/if}
				{/if}

				<div class="flex flex-col gap-1.5">
					<Label>Category</Label>
					<CategoryPicker {familyID} {categories} bind:selectedID={cfCategoryID} />
				</div>
			</div>

			<Dialog.Footer class="gap-2">
				<Button variant="outline" onclick={() => (isOpen = false)}>Cancel</Button>
				<Button onclick={submit} disabled={!cf.title.trim() || (createType === 'event' && !cfEventRange.start) || (createType === 'birthday' && !cfDueDate)}>
					Create
				</Button>
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
