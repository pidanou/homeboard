<script lang="ts">
	import type { Task, CalEvent, Member, AppCategory } from '$lib/types';
	import { calDateToISO, isoToCalDate, fmtCalDate, calDateTimeToISO, taskHasTime, isoToTimeInput } from '$lib/dates';
	import { api } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import * as Select from '$lib/components/ui/select';
	import * as Popover from '$lib/components/ui/popover';
	import { Calendar } from '$lib/components/ui/calendar';
	import type { DateRange } from 'bits-ui';
	import { CalendarDate } from '@internationalized/date';
	import { CalendarDays, User, Users, MapPin, Repeat, Tag, AlignLeft, Trash2 } from 'lucide-svelte';
	import CategoryPicker from '$lib/components/CategoryPicker.svelte';
	import FormRow from '$lib/components/FormRow.svelte';
	import TimePicker from '$lib/components/TimePicker.svelte';
	import * as msg from '$lib/paraglide/messages.js';
	import { getLocale } from '$lib/paraglide/runtime';
	import { buildRRule, parseRRule, type RepeatEndType } from '$lib/rrule';
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
	let efStartPickerOpen = $state(false);
	let efEndPickerOpen = $state(false);
	let efCategoryID = $state<string | undefined>(undefined);
	type RepeatVal = 'none' | 'daily' | 'weekly' | 'monthly' | 'yearly';
	let efRepeat = $state<RepeatVal>('none');
	let efRepeatEndType = $state<RepeatEndType>('never');
	let efRepeatUntil = $state<CalendarDate | undefined>(undefined);
	let efRepeatUntilOpen = $state(false);
	let efRepeatCount = $state(5);
	let efIsRecurring = $state(false);
	let efScopePrompt = $state<'save' | null>(null);
	let efBirthdayOf = $state<string | undefined>(undefined);
	let deleteConfirmOpen = $state(false);

	const REPEAT_LABELS = $derived<Record<string, string>>({
		none: msg.repeat_none(), daily: msg.repeat_daily(), weekly: msg.repeat_weekly(), monthly: msg.repeat_monthly(), yearly: msg.repeat_yearly()
	});
	const REPEAT_END_LABELS = $derived<Record<RepeatEndType, string>>({
		never: msg.repeat_ends_never(), until: msg.repeat_ends_on_date(), count: msg.repeat_ends_after(),
	});

	export function openTask(t: Task, open = true) {
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
		if (open) isOpen = true;
	}

	export function openEvent(e: CalEvent, open = true) {
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
		const parsed = parseRRule(e.recurrence_rule);
		efRepeat = parsed.freq;
		efRepeatEndType = parsed.end.type;
		efRepeatUntil = parsed.end.type === 'until' ? parsed.end.date : undefined;
		efRepeatCount = parsed.end.type === 'count' ? parsed.end.count : 5;
		efIsRecurring = !!e.is_recurring;
		efScopePrompt = null;
		efBirthdayOf = e.birthday_of ?? undefined;
		if (open) isOpen = true;
	}

	function toggleAttendee(ids: string[], uid: string): string[] {
		return ids.includes(uid) ? ids.filter((id) => id !== uid) : [...ids, uid];
	}

	function save() {
		if (efBirthdayOf?.trim()) ef.title = efBirthdayOf.trim() + msg.dialog_birthday_suffix();
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
				const isBirthday = !!efBirthdayOf?.trim();
				const isAllDay = ef.allDay || isBirthday;
				// Birthdays are always single-day; add 1 day for the iCal exclusive-end convention.
				const efEnd = isBirthday ? efEventRange.start.add({ days: 1 }) : (efEventRange.end ?? efEventRange.start);
				await api.patch(`/api/v1/households/${familyID}/events/${id}`, {
					title: ef.title.trim(), description: ef.description, location: ef.location,
					start_at: calDateTimeToISO(efEventRange.start, efStartTime, isAllDay),
					end_at: calDateTimeToISO(efEnd, efEndTime, isAllDay),
					all_day: isAllDay, attendee_ids: ef.attendeeIDs, category_id: efCategoryID,
					important: ef.important,
					recurrence_rule: isBirthday ? 'FREQ=YEARLY' : (buildRRule(efRepeat, efRepeatEndType === 'until' && efRepeatUntil ? { type: 'until', date: efRepeatUntil } : efRepeatEndType === 'count' ? { type: 'count', count: efRepeatCount } : { type: 'never' }) ?? null),
					birthday_of: efBirthdayOf?.trim() || null,
				});
			}
			onSaved();
		} catch { }
	}

	export function deleteTask(t: Task) { openTask(t, false); confirmDelete(); }
	export function deleteEvent(e: CalEvent) { openEvent(e, false); confirmDelete(); }

	function confirmDelete() { deleteConfirmOpen = true; }

	async function doDelete(id: string) {
		isOpen = false;
		deleteConfirmOpen = false;
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
	// German capitalizes all nouns, so lowercasing "Aufgabe"/"Event" mid-sentence would be wrong.
	function lowerKind(kind: string) { return getLocale() === 'de' ? kind : kind.toLowerCase(); }
</script>

<Dialog.Root bind:open={isOpen}>
	<Dialog.Portal>
		<Dialog.Overlay />
		<Dialog.Content class="sm:max-w-md flex flex-col max-h-[90dvh]">
			<Dialog.Header>
				<Dialog.Title class="sr-only">{msg.edit_dialog_title({ kind: lowerKind(editKind === 'task' ? msg.dialog_type_task() : msg.dialog_type_event()) })}</Dialog.Title>
			</Dialog.Header>

			<div class="flex flex-col gap-3 py-2 overflow-y-auto flex-1 min-h-0 px-1"
				onkeydown={(e) => { if (e.key === 'Enter' && (e.target as HTMLElement).tagName !== 'TEXTAREA') { e.preventDefault(); save(); } }}>
				<!-- Title -->
				{#if efBirthdayOf !== undefined}
					<Input bind:value={efBirthdayOf} placeholder={msg.dialog_birthday_name_placeholder()} class="flex-1 h-11 text-lg font-medium" />
				{:else}
				<Input bind:value={ef.title} class="flex-1 h-11 text-lg font-medium" />
				{/if}

				{#if editKind === 'task'}
					<FormRow bind:checked={ef.important}>
						{msg.dialog_important()}
					</FormRow>

					<FormRow icon={CalendarDays}>
						<div class="flex items-center gap-2">
							<Popover.Root bind:open={efDueOpen}>
								<Popover.Trigger class="flex-1">
									<Button variant="outline" class="w-full justify-start font-normal text-sm">
										{efDueDate ? fmtCalDate(efDueDate) : msg.dialog_no_due_date()}
									</Button>
								</Popover.Trigger>
								<Popover.Content class="w-auto p-0" align="start">
									<Calendar type="single" locale={getLocale()} bind:value={efDueDate} onValueChange={() => (efDueOpen = false)} />
								</Popover.Content>
							</Popover.Root>
							{#if efDueDate}
								<TimePicker bind:value={efDueTime} class="w-32 shrink-0" />
							{/if}
						</div>
					</FormRow>

					{#if members.length > 0}
						<FormRow icon={User}>
							<Select.Root type="single" bind:value={ef.assignedTo}>
								<Select.Trigger class="w-full">{members.find(mem => mem.user_id === ef.assignedTo)?.name ?? msg.dialog_unassigned()}</Select.Trigger>
								<Select.Content>
									<Select.Item value="">{msg.dialog_unassigned()}</Select.Item>
									{#each members as mem}
										<Select.Item value={mem.user_id}>{mem.name}</Select.Item>
									{/each}
								</Select.Content>
							</Select.Root>
						</FormRow>
					{/if}

					<FormRow icon={Tag}>
						<CategoryPicker {familyID} {categories} bind:selectedID={efCategoryID} />
					</FormRow>

					<FormRow icon={AlignLeft}>
						<Textarea bind:value={ef.description} placeholder={msg.dialog_notes_placeholder()} rows={2} />
					</FormRow>
				{:else if efBirthdayOf !== undefined}
					<!-- Birthday: single date picker, always all-day, always yearly -->
					<FormRow icon={CalendarDays}>
						<Popover.Root bind:open={efEventPickerOpen}>
							<Popover.Trigger class="w-full">
								<Button variant="outline" class="w-full justify-start font-normal text-sm">
									{efEventRange.start ? fmtCalDate(efEventRange.start) : msg.dialog_birthday_date_placeholder()}
								</Button>
							</Popover.Trigger>
							<Popover.Content class="w-auto p-0" align="start">
								<Calendar
									type="single"
									locale={getLocale()}
									value={efEventRange.start}
									onValueChange={(d) => { efEventRange = { start: d, end: d }; efEventPickerOpen = false; }}
								/>
							</Popover.Content>
						</Popover.Root>
					</FormRow>
				{:else}
					<FormRow icon={CalendarDays}>
						<div class="flex flex-col gap-2">
							<div class="flex items-center gap-2">
								<Popover.Root bind:open={efStartPickerOpen}>
									<Popover.Trigger class="flex-1">
										<Button variant="outline" class="w-full justify-start font-normal text-sm">
											{efEventRange.start ? fmtCalDate(efEventRange.start) : msg.dates_select_dates()}
										</Button>
									</Popover.Trigger>
									<Popover.Content class="w-auto p-0" align="start">
										<Calendar
											type="single"
											locale={getLocale()}
											bind:value={efEventRange.start}
											onValueChange={(d) => {
												if (d && efEventRange.end && efEventRange.end.compare(d) < 0) efEventRange.end = d;
												efStartPickerOpen = false;
											}}
										/>
									</Popover.Content>
								</Popover.Root>
								{#if !ef.allDay}
									<TimePicker bind:value={efStartTime} class="w-28 shrink-0" />
								{/if}
							</div>
							<div class="flex items-center gap-2">
								<Popover.Root bind:open={efEndPickerOpen}>
									<Popover.Trigger class="flex-1">
										<Button variant="outline" class="w-full justify-start font-normal text-sm">
											{efEventRange.end ? fmtCalDate(efEventRange.end) : msg.dates_select_dates()}
										</Button>
									</Popover.Trigger>
									<Popover.Content class="w-auto p-0" align="start">
										<Calendar
											type="single"
											locale={getLocale()}
											bind:value={efEventRange.end}
											minValue={efEventRange.start}
											onValueChange={() => (efEndPickerOpen = false)}
										/>
									</Popover.Content>
								</Popover.Root>
								{#if !ef.allDay}
									<TimePicker bind:value={efEndTime} class="w-28 shrink-0" />
								{/if}
							</div>
						</div>
					</FormRow>

					<FormRow bind:checked={ef.allDay}>
						{msg.dialog_all_day()}
					</FormRow>

					<FormRow bind:checked={ef.important}>
						{msg.dialog_important()}
					</FormRow>

					<FormRow icon={Repeat}>
						<Select.Root type="single" bind:value={efRepeat}>
							<Select.Trigger class="w-full">{REPEAT_LABELS[efRepeat] ?? msg.repeat_none()}</Select.Trigger>
							<Select.Content>
								<Select.Item value="none">{msg.repeat_none()}</Select.Item>
								<Select.Item value="daily">{msg.repeat_daily()}</Select.Item>
								<Select.Item value="weekly">{msg.repeat_weekly()}</Select.Item>
								<Select.Item value="monthly">{msg.repeat_monthly()}</Select.Item>
								<Select.Item value="yearly">{msg.repeat_yearly()}</Select.Item>
							</Select.Content>
						</Select.Root>
					</FormRow>

					{#if efRepeat !== 'none'}
						<FormRow>
							<div class="flex items-center gap-2">
								<span class="text-sm text-muted-foreground shrink-0">{msg.repeat_ends_label()}</span>
								<Select.Root type="single" bind:value={efRepeatEndType}>
									<Select.Trigger class="w-32 shrink-0">{REPEAT_END_LABELS[efRepeatEndType]}</Select.Trigger>
									<Select.Content>
										<Select.Item value="never">{msg.repeat_ends_never()}</Select.Item>
										<Select.Item value="until">{msg.repeat_ends_on_date()}</Select.Item>
										<Select.Item value="count">{msg.repeat_ends_after()}</Select.Item>
									</Select.Content>
								</Select.Root>
								{#if efRepeatEndType === 'until'}
									<Popover.Root bind:open={efRepeatUntilOpen}>
										<Popover.Trigger class="flex-1">
											<Button variant="outline" class="w-full justify-start font-normal text-sm">
												{efRepeatUntil ? fmtCalDate(efRepeatUntil) : msg.dates_select_dates()}
											</Button>
										</Popover.Trigger>
										<Popover.Content class="w-auto p-0" align="start">
											<Calendar type="single" locale={getLocale()} bind:value={efRepeatUntil} onValueChange={() => (efRepeatUntilOpen = false)} />
										</Popover.Content>
									</Popover.Root>
								{:else if efRepeatEndType === 'count'}
									<div class="flex items-center gap-2 flex-1">
										<Input type="number" min="1" bind:value={efRepeatCount} class="w-16" />
										<span class="text-sm text-muted-foreground">{msg.repeat_ends_occurrences()}</span>
									</div>
								{/if}
							</div>
						</FormRow>
					{/if}

					<FormRow icon={MapPin}>
						<Input bind:value={ef.location} placeholder={msg.dialog_location_placeholder()} />
					</FormRow>

					{#if members.length > 0}
						<FormRow icon={Users} compact>
							<div class="flex flex-col gap-1.5">
								{#each members as mem}
									<label class="flex items-center gap-2 text-sm cursor-pointer">
										<Checkbox
											checked={ef.attendeeIDs.includes(mem.user_id)}
											onCheckedChange={() => (ef.attendeeIDs = toggleAttendee(ef.attendeeIDs, mem.user_id))}
										/>
										{mem.name}
									</label>
								{/each}
							</div>
						</FormRow>
					{/if}

					<FormRow icon={Tag}>
						<CategoryPicker {familyID} {categories} bind:selectedID={efCategoryID} />
					</FormRow>

					<FormRow icon={AlignLeft}>
						<Textarea bind:value={ef.description} placeholder={msg.dialog_notes_placeholder()} rows={2} />
					</FormRow>
				{/if}
			</div>

			<Dialog.Footer class="flex-col gap-2">
				{#if efScopePrompt}
					<p class="text-sm text-muted-foreground text-center">
						{msg.edit_dialog_update_which()}
					</p>
					<div class="flex gap-2 justify-center">
						<Button variant="outline" size="sm" onclick={() => doSave(editID)}>{msg.edit_dialog_this_event()}</Button>
						<Button variant="outline" size="sm" onclick={() => doSave(parentID(editID))}>{msg.edit_dialog_all_events()}</Button>
						<Button variant="ghost" size="sm" onclick={() => (efScopePrompt = null)}>{msg.dialog_cancel()}</Button>
					</div>
				{:else}
					<div class="flex w-full items-center gap-2">
						<Button
							variant="outline"
							size="icon"
							class="text-destructive hover:text-destructive shrink-0"
							onclick={confirmDelete}
							aria-label={msg.edit_dialog_delete()}
						>
							<Trash2 class="size-4" />
						</Button>
						<div class="flex flex-1 justify-end gap-2">
							<Button variant="outline" class="flex-1 sm:flex-none" onclick={() => (isOpen = false)}>{msg.dialog_cancel()}</Button>
							<Button
								class="flex-1 sm:flex-none"
								onclick={save}
								disabled={!ef.title.trim() || (editKind === 'event' && !efEventRange.start)}
							>{msg.edit_dialog_save()}</Button>
						</div>
					</div>
				{/if}
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<AlertDialog.Root bind:open={deleteConfirmOpen}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>{msg.edit_dialog_delete_confirm({ title: ef.title })}</AlertDialog.Title>
			{#if editKind === 'event' && efIsRecurring && !efBirthdayOf?.trim()}
				<AlertDialog.Description>{msg.edit_dialog_delete_which()}</AlertDialog.Description>
			{/if}
		</AlertDialog.Header>
		<AlertDialog.Footer>
			{#if editKind === 'event' && efIsRecurring && !efBirthdayOf?.trim()}
				<Button variant="outline" size="sm" onclick={() => (deleteConfirmOpen = false)}>{msg.dialog_cancel()}</Button>
				<Button variant="destructive" size="sm" onclick={() => doDelete(editID)}>{msg.edit_dialog_this_event()}</Button>
				<Button variant="destructive" size="sm" onclick={() => doDelete(parentID(editID))}>{msg.edit_dialog_all_events()}</Button>
			{:else}
				<AlertDialog.Cancel>{msg.dialog_cancel()}</AlertDialog.Cancel>
				<Button variant="destructive" onclick={() => doDelete(efBirthdayOf?.trim() ? parentID(editID) : editID)}>{msg.edit_dialog_delete()}</Button>
			{/if}
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
