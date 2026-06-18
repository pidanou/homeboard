// Svelte action for vertical list drag-to-reorder via Pointer Events (touch + mouse).
// Usage: <ul use:sortable={{ onReorder }}>  — each child must have data-id attribute.

export type ReorderCallback = (ids: string[]) => void;

export function sortable(node: HTMLElement, options: { onReorder: ReorderCallback }) {
	let dragEl: HTMLElement | null = null;
	let placeholder: HTMLElement | null = null;
	let startY = 0;
	let offsetY = 0;

	function getItems(): HTMLElement[] {
		return Array.from(node.children).filter(
			(el) => el !== placeholder
		) as HTMLElement[];
	}

	function onPointerDown(e: PointerEvent) {
		const handle = (e.target as HTMLElement).closest('[data-drag-handle]');
		if (!handle) return;
		const target = (e.target as HTMLElement).closest('[data-id]') as HTMLElement | null;
		if (!target || !node.contains(target)) return;

		dragEl = target;
		const rect = dragEl.getBoundingClientRect();
		startY = e.clientY;
		offsetY = e.clientY - rect.top;

		placeholder = document.createElement('div');
		placeholder.style.height = `${rect.height}px`;
		placeholder.className = 'rounded-lg bg-muted/50 border border-dashed border-border';
		dragEl.after(placeholder);

		dragEl.style.position = 'fixed';
		dragEl.style.left = `${rect.left}px`;
		dragEl.style.width = `${rect.width}px`;
		dragEl.style.top = `${rect.top}px`;
		dragEl.style.zIndex = '50';
		dragEl.style.opacity = '0.9';
		dragEl.style.pointerEvents = 'none';
		dragEl.style.boxShadow = '0 4px 12px rgba(0,0,0,0.15)';

		node.setPointerCapture(e.pointerId);
		e.preventDefault();
	}

	function onPointerMove(e: PointerEvent) {
		if (!dragEl || !placeholder) return;

		const dy = e.clientY - startY;
		const rect = dragEl.getBoundingClientRect();
		dragEl.style.top = `${rect.top + dy}px`;
		startY = e.clientY;

		const midY = e.clientY - offsetY + rect.height / 2;
		const items = getItems();
		let inserted = false;
		for (const item of items) {
			const r = item.getBoundingClientRect();
			if (midY < r.top + r.height / 2) {
				node.insertBefore(placeholder, item);
				inserted = true;
				break;
			}
		}
		if (!inserted) node.appendChild(placeholder);
	}

	function onPointerUp() {
		if (!dragEl || !placeholder) return;

		dragEl.style.cssText = '';
		placeholder.replaceWith(dragEl);
		placeholder = null;

		const ids = getItems().map((el) => el.dataset.id!);
		options.onReorder(ids);
		dragEl = null;
	}

	node.addEventListener('pointerdown', onPointerDown);
	node.addEventListener('pointermove', onPointerMove);
	node.addEventListener('pointerup', onPointerUp);
	node.addEventListener('pointercancel', onPointerUp);

	return {
		destroy() {
			node.removeEventListener('pointerdown', onPointerDown);
			node.removeEventListener('pointermove', onPointerMove);
			node.removeEventListener('pointerup', onPointerUp);
			node.removeEventListener('pointercancel', onPointerUp);
		},
		update(newOptions: { onReorder: ReorderCallback }) {
			options = newOptions;
		},
	};
}
