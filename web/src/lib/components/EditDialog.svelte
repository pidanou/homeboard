<script lang="ts">
	import type { Task, CalEvent, Member, AppCategory } from '$lib/types';
	import { calDateToISO, isoToCalDate, fmtCalDate, calDateTimeToISO, rangeLabelFor } from '$lib/dates';
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
	import { CalendarDate } from '@internationalized/date';
	import { CalendarDays, Repeat } from 'lucide-svelte';
	import CategoryPicker from '$lib/components/CategoryPicker.svelte';
	import IconPicker from '$lib/components/IconPicker.svelte';

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
	let efEventRange = $state<DateRange>({ start: undefined, end: undefined });
	let efStartTime = $state('09:00');
	let efEndTime = $state('10:00');
	let efEventPickerOpen = $state(false);
	let efCategoryID = $state<string | undefined>(undefined);
	type RepeatVal = 'none' | 'daily' | 'weekly' | 'monthly' | 'yearly';
	let efRepeat = $state<RepeatVal>('none');
	let efIsRecurring = $state(false);
	let efScopePrompt = $state<'save' | 'delete' | null>(null);
	let efIcon = $state<string | undefined>(undefined);
	let efBirthdayOf = $state<string | undefined>(undefined);

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
		efCategoryID = t.category_id;
		efIcon = t.icon;
		isOpen = true;
	}

	export function openEvent(e: CalEvent) {
		editKind = 'event';
		editID = e.id;
		ef = {
			title: e.title, description: e.description ?? '', location: e.location ?? '',
			allDay: e.all_day, important: false, status: '',
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
		efIcon = e.icon;
		efBirthdayOf = e.birthday_of ?? undefined;
		isOpen = true;
	}

	function toggleAttendee(ids: string[], uid: string): string[] {
		return ids.includes(uid) ? ids.filter((id) => id !== uid) : [...ids, uid];
	}

	function save() {
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
					end_date: efDueDate ? calDateToISO(efDueDate) : undefined,
					category_id: efCategoryID,
					icon: efIcon ?? null,
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
					recurrence_rule: efBirthdayOf?.trim() ? RRULE['yearly'] : (efRepeat !== 'none' ? RRULE[efRepeat] : null),
					icon: efIcon ?? null,
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

			<div class="flex flex-col gap-4 py-2 overflow-y-auto flex-1 min-h-0 px-1">
				<div class="flex flex-col gap-1.5">
					<Label for="ef-title">Title</Label>
					<div class="flex gap-2">
						<IconPicker bind:value={efIcon} />
						<Input
							id="ef-title"
							bind:value={ef.title}
							class="flex-1"
							onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); save(); } }}
						/>
					</div>
				</div>

				<div class="flex flex-col gap-1.5">
					<Label for="ef-desc">Description</Label>
					<Textarea id="ef-desc" bind:value={ef.description} placeholder="Optional details…" rows={2} />
				</div>

				{#if editKind === 'task'}
					<div class="flex gap-3">
						<div class="flex flex-col gap-1.5 flex-1">
							<label class="flex items-center gap-2 text-sm cursor-pointer mt-5">
								<Checkbox bind:checked={ef.important} />
								Important
							</label>
						</div>
						<div class="flex flex-col gap-1.5 flex-1">
							<Label>Due date</Label>
							<Popover.Root bind:open={efDueOpen}>
								<Popover.Trigger>
									<Button variant="outline" class="w-full justify-start gap-2 font-normal text-sm">
										<CalendarDays class="w-4 h-4 text-muted-foreground shrink-0" />
										{efDueDate ? fmtCalDate(efDueDate) : 'Pick a date'}
									</Button>
								</Popover.Trigger>
								<Popover.Content class="w-auto p-0" align="start">
									<Calendar type="single" bind:value={efDueDate} onValueChange={() => (efDueOpen = false)} />
								</Popover.Content>
							</Popover.Root>
						</div>
					</div>
					{#if members.length > 0}
						<div class="flex flex-col gap-1.5">
							<Label>Assign to</Label>
							<Select.Root type="single" bind:value={ef.assignedTo}>
								<Select.Trigger class="w-full">{members.find(m => m.user_id === ef.assignedTo)?.name ?? 'Unassigned'}</Select.Trigger>
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
					</div>
					{#if !ef.allDay}
						<div class="flex gap-3">
							<div class="flex flex-col gap-1.5 flex-1">
								<Label for="ef-start-time">Start time</Label>
								<Input id="ef-start-time" type="time" bind:value={efStartTime} />
							</div>
							<div class="flex flex-col gap-1.5 flex-1">
								<Label for="ef-end-time">End time</Label>
								<Input id="ef-end-time" type="time" bind:value={efEndTime} />
							</div>
						</div>
					{/if}
					<label class="flex items-center gap-2 text-sm cursor-pointer">
						<Checkbox bind:checked={ef.allDay} />
						All day
					</label>
					<div class="flex flex-col gap-1.5">
						<Label>Repeat</Label>
						<Select.Root type="single" bind:value={efRepeat}>
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
						<Label for="ef-location">Location</Label>
						<Input id="ef-location" bind:value={ef.location} placeholder="Optional location…" />
					</div>
					<div class="flex flex-col gap-1.5">
						<Label for="ef-birthday-of">Birthday of</Label>
						<Input id="ef-birthday-of" bind:value={efBirthdayOf} placeholder="Name (sets yearly recurrence)" />
					</div>
					{#if members.length > 0}
						<div class="flex flex-col gap-1.5">
							<Label>Members</Label>
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
						</div>
					{/if}
				{/if}

				<div class="flex flex-col gap-1.5">
					<Label>Category</Label>
					<CategoryPicker {familyID} {categories} bind:selectedID={efCategoryID} />
				</div>
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
