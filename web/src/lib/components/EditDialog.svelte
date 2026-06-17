<script lang="ts">
	import type { Task, CalEvent, Member, AppLabel } from '$lib/types';
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
	import { CalendarDays } from 'lucide-svelte';
	import LabelPicker from '$lib/components/LabelPicker.svelte';

	let { familyID, members, labels, onSaved, onDeleted, onError }: {
		familyID: string;
		members: Member[];
		labels: AppLabel[];
		onSaved: () => void;
		onDeleted: () => void;
		onError: (e: unknown) => void;
	} = $props();

	let isOpen = $state(false);
	let editKind = $state<'task' | 'event'>('task');
	let editID = $state('');
	let ef = $state({
		title: '', description: '', priority: 'medium', status: 'todo',
		allDay: false, location: '', assignedTo: '', attendeeIDs: [] as string[],
	});
	let efDueDate = $state<CalendarDate | undefined>(undefined);
	let efDueOpen = $state(false);
	let efEventRange = $state<DateRange>({ start: undefined, end: undefined });
	let efStartTime = $state('09:00');
	let efEndTime = $state('10:00');
	let efEventPickerOpen = $state(false);
	let efLabelIDs = $state<string[]>([]);

	export function openTask(t: Task) {
		editKind = 'task';
		editID = t.id;
		ef = {
			title: t.title, description: t.description ?? '',
			priority: t.priority || 'medium', status: t.status,
			allDay: false, location: '', assignedTo: t.assigned_to ?? '', attendeeIDs: [],
		};
		efDueDate = t.end_date ? isoToCalDate(t.end_date) : undefined;
		efLabelIDs = [...(t.label_ids ?? [])];
		isOpen = true;
	}

	export function openEvent(e: CalEvent) {
		editKind = 'event';
		editID = e.id;
		ef = {
			title: e.title, description: e.description ?? '', location: e.location ?? '',
			allDay: e.all_day, priority: 'medium', status: '',
			assignedTo: '', attendeeIDs: e.attendee_ids ?? [],
		};
		efEventRange = { start: isoToCalDate(e.start_at), end: isoToCalDate(e.end_at) };
		const s = new Date(e.start_at);
		const en = new Date(e.end_at);
		efStartTime = `${String(s.getHours()).padStart(2, '0')}:${String(s.getMinutes()).padStart(2, '0')}`;
		efEndTime = `${String(en.getHours()).padStart(2, '0')}:${String(en.getMinutes()).padStart(2, '0')}`;
		efLabelIDs = [...(e.label_ids ?? [])];
		isOpen = true;
	}

	function toggleAttendee(ids: string[], uid: string): string[] {
		return ids.includes(uid) ? ids.filter((id) => id !== uid) : [...ids, uid];
	}

	async function save() {
		if (!ef.title.trim()) return;
		isOpen = false;
		try {
			if (editKind === 'task') {
				await api.patch(`/api/v1/families/${familyID}/tasks/${editID}`, {
					title: ef.title.trim(), description: ef.description,
					priority: ef.priority, status: ef.status,
					assigned_to: ef.assignedTo || undefined,
					end_date: efDueDate ? calDateToISO(efDueDate) : undefined,
					label_ids: efLabelIDs,
				});
			} else {
				if (!efEventRange.start) return;
				const efEnd = efEventRange.end ?? efEventRange.start;
				await api.patch(`/api/v1/families/${familyID}/events/${editID}`, {
					title: ef.title.trim(), description: ef.description, location: ef.location,
					start_at: calDateTimeToISO(efEventRange.start, efStartTime, ef.allDay),
					end_at: calDateTimeToISO(efEnd, efEndTime, ef.allDay),
					all_day: ef.allDay, attendee_ids: ef.attendeeIDs, label_ids: efLabelIDs,
				});
			}
			onSaved();
		} catch (e) {
			onError(e);
		}
	}

	async function del() {
		isOpen = false;
		try {
			if (editKind === 'task') {
				await api.delete(`/api/v1/families/${familyID}/tasks/${editID}`);
			} else {
				await api.delete(`/api/v1/families/${familyID}/events/${editID}`);
			}
			onDeleted();
		} catch (e) {
			onError(e);
		}
	}
</script>

<Dialog.Root bind:open={isOpen}>
	<Dialog.Portal>
		<Dialog.Overlay />
		<Dialog.Content class="sm:max-w-md">
			<Dialog.Header>
				<Dialog.Title>Edit {editKind}</Dialog.Title>
			</Dialog.Header>

			<div class="flex flex-col gap-4 py-2">
				<div class="flex flex-col gap-1.5">
					<Label for="ef-title">Title</Label>
					<Input
						id="ef-title"
						bind:value={ef.title}
						onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); save(); } }}
					/>
				</div>

				<div class="flex flex-col gap-1.5">
					<Label for="ef-desc">Description</Label>
					<Textarea id="ef-desc" bind:value={ef.description} placeholder="Optional details…" rows={2} />
				</div>

				{#if editKind === 'task'}
					<div class="flex gap-3">
						<div class="flex flex-col gap-1.5 flex-1">
							<Label>Priority</Label>
							<Select.Root type="single" bind:value={ef.priority}>
								<Select.Trigger class="w-full"><SelectPrimitive.Value /></Select.Trigger>
								<Select.Content>
									<Select.Item value="low">Low</Select.Item>
									<Select.Item value="medium">Medium</Select.Item>
									<Select.Item value="high">High</Select.Item>
								</Select.Content>
							</Select.Root>
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
						<Label for="ef-location">Location</Label>
						<Input id="ef-location" bind:value={ef.location} placeholder="Optional location…" />
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
					<Label>Labels</Label>
					<LabelPicker {familyID} {labels} bind:selectedIDs={efLabelIDs} onError={onError} />
				</div>
			</div>

			<Dialog.Footer class="flex-col-reverse sm:flex-row gap-2">
				<Button variant="destructive" onclick={del}>Delete</Button>
				<div class="flex gap-2 sm:ml-auto">
					<Button variant="outline" onclick={() => (isOpen = false)}>Cancel</Button>
					<Button
						onclick={save}
						disabled={!ef.title.trim() || (editKind === 'event' && !efEventRange.start)}
					>Save</Button>
				</div>
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
