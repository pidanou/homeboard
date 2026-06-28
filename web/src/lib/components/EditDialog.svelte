<script lang="ts">
	import type { Task, CalEvent, Member, AppCategory } from '$lib/types';
	import { calDateToISO, isoToCalDate, fmtCalDate, calDateTimeToISO, rangeLabelFor, taskHasTime, isoToTimeInput } from '$lib/dates';
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
	import { CalendarDate } from '@internationalized/date';
	import { CalendarDays } from 'lucide-svelte';
	import CategoryPicker from '$lib/components/CategoryPicker.svelte';
	let { familyID, members, categories, onSaved, onDeleted }: {
		familyID: string;
		members: Member[];
		categories: AppCategory[];
		onSaved: () => void;
		onDeleted: () => void;
	} = $props();

	let isOpen = $state(false);
	let editKind = $state<'task' | 'event'>('task');
	let editID = $state('');
	let ef = $state({
		title: '', description: '', important: false, status: 'todo',
		allDay: false, location: '', assignedTo: '', attendeeIDs: [] as string[],
	});
	let efDueDate = $state<CalendarDate | undefined>(undefined);
	let efDueOpen = $state(false);
	let efDueTime = $state('');
	let efEventRange = $state<DateRange>({ start: undefined, end: undefined });
	let efStartTime = $state('09:00');
	let efEndTime = $state('10:00');
	let efEventPickerOpen = $state(false);
	let efCategoryID = $state<string | undefined>(undefined);
	type RepeatVal = 'none' | 'daily' | 'weekly' | 'monthly' | 'yearly';
	let efRepeat = $state<RepeatVal>('none');
	let efIsRecurring = $state(false);
	let efScopePrompt = $state<'save' | 'delete' | null>(null);
	let efBirthdayOf = $state<string | undefined>(undefined);
	let efShowMore = $state(false);

	const REPEAT_LABELS: Record<string, string> = {
		none: 'Does not repeat', daily: 'Daily', weekly: 'Weekly', monthly: 'Monthly', yearly: 'Yearly'
	};

	const RRULE: Record<string, string> = {
		daily: 'FREQ=DAILY', weekly: 'FREQ=WEEKLY', monthly: 'FREQ=MONTHLY', yearly: 'FREQ=YEARLY',
	};
	const RRULE_REVERSE: Record<string, string> = {
		'FREQ=DAILY': 'daily', 'FREQ=WEEKLY': 'weekly', 'FREQ=MONTHLY': 'monthly', 'FREQ=YEARLY': 'yearly',
	};

	export function openTask(t: Task) {
		editKind = 'task';
		editID = t.id;
		ef = {
			title: t.title, description: t.description ?? '',
			important: t.important ?? false, status: t.status,
			allDay: false, location: '', assignedTo: t.assigned_to ?? '', attendeeIDs: [],
		};
		efDueDate = t.end_date ? isoToCalDate(t.end_date) : undefined;
		efDueTime = t.end_date && taskHasTime(t.end_date) ? isoToTimeInput(t.end_date) : '';
		efCategoryID = t.category_id;
		efBirthdayOf = undefined;
		efShowMore = !!(t.description || t.category_id);
		isOpen = true;
	}

	export function openEvent(e: CalEvent) {
		editKind = 'event';
		editID = e.id;
		ef = {
			title: e.title, description: e.description ?? '', location: e.location ?? '',
			allDay: e.all_day, important: e.important ?? false, status: '',
			assignedTo: '', attendeeIDs: e.attendee_ids ?? [],
		};
		efEventRange = { start: isoToCalDate(e.start_at), end: isoToCalDate(e.end_at) };
		const s = new Date(e.start_at);
		const en = new Date(e.end_at);
		efStartTime = `${String(s.getHours()).padStart(2, '0')}:${String(s.getMinutes()).padStart(2, '0')}`;
		efEndTime = `${String(en.getHours()).padStart(2, '0')}:${String(en.getMinutes()).padStart(2, '0')}`;
		efCategoryID = e.category_id;
		efRepeat = (e.recurrence_rule ? (RRULE_REVERSE[e.recurrence_rule] ?? 'none') : 'none') as RepeatVal;
		efIsRecurring = !!e.is_recurring;
		efScopePrompt = null;
		efBirthdayOf = e.birthday_of ?? undefined;
		efShowMore = !!(e.description || e.recurrence_rule || e.location || e.birthday_of || e.category_id);
		isOpen = true;
	}

	function toggleAttendee(ids: string[], uid: string): string[] {
		return ids.includes(uid) ? ids.filter((id) => id !== uid) : [...ids, uid];
	}

	function save() {
		if (efBirthdayOf?.trim()) ef.title = efBirthdayOf.trim() + "'s Birthday";
		if (!ef.title.trim()) return;
		if (editKind === 'event' && efIsRecurring) { efScopePrompt = 'save'; return; }
		doSave(editID);
	}

	async function doSave(id: string) {
		isOpen = false;
		try {
			if (editKind === 'task') {
				await api.patch(`/api/v1/households/${familyID}/tasks/${id}`, {
					title: ef.title.trim(), description: ef.description,
					important: ef.important, status: ef.status,
					assigned_to: ef.assignedTo || undefined,
					end_date: efDueDate ? (efDueTime ? calDateTimeToISO(efDueDate, efDueTime, false) : calDateToISO(efDueDate)) : undefined,
					category_id: efCategoryID,
					birthday_of: efBirthdayOf ?? null,
				});
			} else {
				if (!efEventRange.start) return;
				const efEnd = efEventRange.end ?? efEventRange.start;
				await api.patch(`/api/v1/households/${familyID}/events/${id}`, {
					title: ef.title.trim(), description: ef.description, location: ef.location,
					start_at: calDateTimeToISO(efEventRange.start, efStartTime, ef.allDay),
					end_at: calDateTimeToISO(efEnd, efEndTime, ef.allDay),
					all_day: ef.allDay, attendee_ids: ef.attendeeIDs, category_id: efCategoryID,
					important: ef.important,
					recurrence_rule: efBirthdayOf?.trim() ? RRULE['yearly'] : (efRepeat !== 'none' ? RRULE[efRepeat] : null),
					birthday_of: efBirthdayOf?.trim() || null,
				});
			}
			onSaved();
		} catch { }
	}

	function del() {
		if (editKind === 'event' && efIsRecurring) { efScopePrompt = 'delete'; return; }
		doDelete(editID);
	}

	async function doDelete(id: string) {
		isOpen = false;
		try {
			if (editKind === 'task') {
				await api.delete(`/api/v1/households/${familyID}/tasks/${id}`);
			} else {
				await api.delete(`/api/v1/households/${familyID}/events/${id}`);
			}
			onDeleted();
		} catch { }
	}

	// Strip the ::YYYYMMDD suffix to get the parent ID.
	function parentID(id: string) { return id.split('::')[0]; }
</script>

<Dialog.Root bind:open={isOpen}>
	<Dialog.Portal>
		<Dialog.Overlay />
		<Dialog.Content class="sm:max-w-md flex flex-col max-h-[90dvh]">
			<Dialog.Header>
				<Dialog.Title>Edit {editKind}</Dialog.Title>
			</Dialog.Header>

			<div class="flex flex-col gap-3 py-2 overflow-y-auto flex-1 min-h-0 px-1"
				onkeydown={(e) => { if (e.key === 'Enter' && (e.target as HTMLElement).tagName !== 'TEXTAREA') { e.preventDefault(); save(); } }}>
				<!-- Title -->
				{#if efBirthdayOf !== undefined}
					<Input bind:value={efBirthdayOf} placeholder="Person's name…" class="flex-1" />
				{:else}
				<Input bind:value={ef.title} class="flex-1" />
				{/if}

				{#if editKind === 'task'}
					<!-- Task primary: important + due date -->
					<div class="flex items-center gap-2">
						<label class="flex items-center gap-2 text-sm cursor-pointer shrink-0">
							<Checkbox bind:checked={ef.important} />
							Important
						</label>
						<Popover.Root bind:open={efDueOpen}>
							<Popover.Trigger class="flex-1">
								<Button variant="outline" class="w-full justify-start gap-2 font-normal text-sm">
									<CalendarDays class="w-4 h-4 text-muted-foreground shrink-0" />
									{efDueDate ? fmtCalDate(efDueDate) : 'No due date'}
								</Button>
							</Popover.Trigger>
							<Popover.Content class="w-auto p-0" align="start">
								<Calendar type="single" bind:value={efDueDate} onValueChange={() => (efDueOpen = false)} />
							</Popover.Content>
						</Popover.Root>
						{#if efDueDate}
							<Input type="time" bind:value={efDueTime} class="w-32 shrink-0" />
						{/if}
					</div>

					<!-- Task secondary -->
					{#if efShowMore}
						{#if members.length > 0}
							<Select.Root type="single" bind:value={ef.assignedTo}>
								<Select.Trigger class="w-full">{members.find(m => m.user_id === ef.assignedTo)?.name ?? 'Unassigned'}</Select.Trigger>
								<Select.Content>
									<Select.Item value="">Unassigned</Select.Item>
									{#each members as m}
										<Select.Item value={m.user_id}>{m.name}</Select.Item>
									{/each}
								</Select.Content>
							</Select.Root>
						{/if}
						<Textarea bind:value={ef.description} placeholder="Notes…" rows={2} />
						<CategoryPicker {familyID} {categories} bind:selectedID={efCategoryID} />
					{/if}
				{:else}
					<!-- Event primary: dates, all day, times -->
					<Popover.Root bind:open={efEventPickerOpen}>
						<Popover.Trigger>
							<Button variant="outline" class="w-full justify-start gap-2 font-normal text-sm">
								<CalendarDays class="w-4 h-4 text-muted-foreground shrink-0" />
								{rangeLabelFor(efEventRange)}
							</Button>
						</Popover.Trigger>
						<Popover.Content class="w-auto p-0" align="start">
							<RangeCalendar
								bind:value={efEventRange}
								onValueChange={() => { if (efEventRange.start && efEventRange.end) efEventPickerOpen = false; }}
							/>
						</Popover.Content>
					</Popover.Root>
					<label class="flex items-center gap-2 text-sm cursor-pointer">
						<Checkbox bind:checked={ef.allDay} />
						All day
					</label>
					<label class="flex items-center gap-2 text-sm cursor-pointer">
						<Checkbox bind:checked={ef.important} />
						Important
					</label>
					{#if !ef.allDay}
						<div class="flex gap-2">
							<Input type="time" bind:value={efStartTime} class="flex-1" />
							<Input type="time" bind:value={efEndTime} class="flex-1" />
						</div>
					{/if}

					<!-- Event secondary -->
					{#if efShowMore}
						<Select.Root type="single" bind:value={efRepeat}>
							<Select.Trigger class="w-full">{REPEAT_LABELS[efRepeat] ?? 'Does not repeat'}</Select.Trigger>
							<Select.Content>
								<Select.Item value="none">Does not repeat</Select.Item>
								<Select.Item value="daily">Daily</Select.Item>
								<Select.Item value="weekly">Weekly</Select.Item>
								<Select.Item value="monthly">Monthly</Select.Item>
								<Select.Item value="yearly">Yearly</Select.Item>
							</Select.Content>
						</Select.Root>
						<Input bind:value={ef.location} placeholder="Location…" />
						{#if members.length > 0}
							<div class="flex flex-col gap-1.5">
								{#each members as m}
									<label class="flex items-center gap-2 text-sm cursor-pointer">
										<Checkbox
											checked={ef.attendeeIDs.includes(m.user_id)}
											onCheckedChange={() => (ef.attendeeIDs = toggleAttendee(ef.attendeeIDs, m.user_id))}
										/>
										{m.name}
									</label>
								{/each}
							</div>
						{/if}
						<Textarea bind:value={ef.description} placeholder="Notes…" rows={2} />
						<CategoryPicker {familyID} {categories} bind:selectedID={efCategoryID} />
					{/if}
				{/if}

				<button
					class="text-xs text-muted-foreground hover:text-foreground transition-colors text-left w-fit"
					onclick={() => (efShowMore = !efShowMore)}
				>{efShowMore ? '− Less options' : '+ More options'}</button>
			</div>

			<Dialog.Footer class="flex-col gap-2">
				{#if efScopePrompt}
					<p class="text-sm text-muted-foreground text-center">
						{efScopePrompt === 'save' ? 'Update which events?' : 'Delete which events?'}
					</p>
					<div class="flex gap-2 justify-center">
						<Button variant="outline" size="sm" onclick={() => {
							if (efScopePrompt === 'save') doSave(editID);
							else doDelete(editID);
						}}>This event</Button>
						<Button variant="outline" size="sm" onclick={() => {
							const pid = parentID(editID);
							if (efScopePrompt === 'save') doSave(pid);
							else doDelete(pid);
						}}>All events</Button>
						<Button variant="ghost" size="sm" onclick={() => (efScopePrompt = null)}>Cancel</Button>
					</div>
				{:else}
					<div class="flex flex-col-reverse sm:flex-row gap-2">
						<Button variant="destructive" onclick={del}>Delete</Button>
						<div class="flex gap-2 sm:ml-auto">
							<Button variant="outline" onclick={() => (isOpen = false)}>Cancel</Button>
							<Button
								onclick={save}
								disabled={!ef.title.trim() || (editKind === 'event' && !efEventRange.start)}
							>Save</Button>
						</div>
					</div>
				{/if}
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
