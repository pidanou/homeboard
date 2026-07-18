<script lang="ts">
	import { Dialog as DialogPrimitive } from "bits-ui";
	import DialogPortal from "./dialog-portal.svelte";
	import type { Snippet } from "svelte";
	import * as Dialog from "./index.js";
	import { cn, type WithoutChildrenOrChild } from "$lib/utils.js";
	import type { ComponentProps } from "svelte";
	import { Button } from "$lib/components/ui/button/index.js";
	import { HugeiconsIcon } from "@hugeicons/svelte"
	import { Cancel01Icon } from '@hugeicons/core-free-icons';

	let {
		ref = $bindable(null),
		class: className,
		portalProps,
		children,
		showCloseButton = true,
		...restProps
	}: WithoutChildrenOrChild<DialogPrimitive.ContentProps> & {
		portalProps?: WithoutChildrenOrChild<ComponentProps<typeof DialogPortal>>;
		children: Snippet;
		showCloseButton?: boolean;
	} = $props();

	let dragX = $state(0);
	let dragY = $state(0);
	let dragging = $state(false);

	function onDragStart(e: PointerEvent) {
		if (window.matchMedia('(pointer: coarse)').matches) return;
		const target = e.target as HTMLElement;
		if (!target.closest('[data-slot="dialog-header"]')) return;
		if (target.closest('button, a, input, textarea, select')) return;
		dragging = true;
		const startX = e.clientX - dragX;
		const startY = e.clientY - dragY;
		const onMove = (ev: PointerEvent) => {
			dragX = ev.clientX - startX;
			dragY = ev.clientY - startY;
		};
		const onUp = () => {
			dragging = false;
			window.removeEventListener('pointermove', onMove);
			window.removeEventListener('pointerup', onUp);
		};
		window.addEventListener('pointermove', onMove);
		window.addEventListener('pointerup', onUp);
	}
</script>

<DialogPortal {...portalProps}>
	<Dialog.Overlay />
	<DialogPrimitive.Content
		bind:ref
		data-slot="dialog-content"
		onpointerdown={onDragStart}
		class={cn(
			"bg-popover text-popover-foreground data-open:animate-in data-closed:animate-out data-closed:fade-out-0 data-open:fade-in-0 data-closed:zoom-out-95 data-open:zoom-in-95 ring-foreground/5 grid max-w-[calc(100%-2rem)] gap-6 rounded-4xl p-6 text-sm ring-1 shadow-2xl shadow-black/30 duration-100 sm:max-w-md fixed top-1/2 left-1/2 z-50 w-full outline-none",
			dragging && "select-none",
			className
		)}
		style="{dragging ? 'transition: none;' : ''}translate: calc(-50% + {dragX}px) calc(-50% + {dragY}px)"
		{...restProps}
	>
		{@render children?.()}
		{#if showCloseButton}
			<DialogPrimitive.Close data-slot="dialog-close">
				{#snippet child({ props })}
					<Button variant="ghost" class="absolute top-4 right-4" size="icon-sm" {...props}>
						<HugeiconsIcon icon={Cancel01Icon} strokeWidth={2}  />
						<span class="sr-only">Close</span>
					</Button>
				{/snippet}
			</DialogPrimitive.Close>
		{/if}
	</DialogPrimitive.Content>
</DialogPortal>
