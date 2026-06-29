<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';

	const W = 640, H = 360; // 16:9 display canvas

	let {
		open = $bindable(false),
		onconfirm
	}: {
		open: boolean;
		onconfirm: (blob: Blob, originalFile?: File) => void;
	} = $props();

	let canvas: HTMLCanvasElement;
	let img = new Image();
	let imgLoaded = $state(false);
	let loadedFile: File | undefined;

	let scale = $state(1);
	let offsetX = $state(0);
	let offsetY = $state(0);

	let dragging = false;
	let lastX = 0;
	let lastY = 0;
	let lastPinchDist = 0;

	export function loadFile(file: File) {
		loadedFile = file;
		const reader = new FileReader();
		reader.onload = (e) => {
			img.onload = () => {
				imgLoaded = true;
				scale = Math.min(img.naturalWidth / W, img.naturalHeight / H);
				offsetX = (img.naturalWidth - W * scale) / 2;
				offsetY = (img.naturalHeight - H * scale) / 2;
				draw();
			};
			img.src = e.target!.result as string;
		};
		reader.readAsDataURL(file);
	}

	function draw() {
		if (!canvas || !imgLoaded) return;
		const ctx = canvas.getContext('2d')!;
		ctx.clearRect(0, 0, W, H);
		ctx.drawImage(img, offsetX, offsetY, W * scale, H * scale, 0, 0, W, H);
	}

	function clampOffset() {
		offsetX = Math.max(0, Math.min(offsetX, img.naturalWidth - W * scale));
		offsetY = Math.max(0, Math.min(offsetY, img.naturalHeight - H * scale));
	}

	function onMouseDown(e: MouseEvent) { dragging = true; lastX = e.clientX; lastY = e.clientY; }
	function onMouseMove(e: MouseEvent) {
		if (!dragging) return;
		offsetX -= (e.clientX - lastX) * scale;
		offsetY -= (e.clientY - lastY) * scale;
		lastX = e.clientX; lastY = e.clientY;
		clampOffset(); draw();
	}
	function onMouseUp() { dragging = false; }

	function zoomBy(ratio: number) {
		const maxScale = Math.min(img.naturalWidth / W, img.naturalHeight / H);
		const minScale = maxScale / 5;
		const newScale = Math.max(minScale, Math.min(scale * ratio, maxScale));
		offsetX += (W / 2) * (scale - newScale);
		offsetY += (H / 2) * (scale - newScale);
		scale = newScale;
		clampOffset(); draw();
	}

	function onWheel(e: WheelEvent) { e.preventDefault(); zoomBy(1 + (e.deltaY > 0 ? 0.03 : -0.03)); }

	function onTouchStart(e: TouchEvent) {
		if (e.touches.length === 1) { dragging = true; lastX = e.touches[0].clientX; lastY = e.touches[0].clientY; }
		else if (e.touches.length === 2) {
			dragging = false;
			lastPinchDist = Math.hypot(e.touches[0].clientX - e.touches[1].clientX, e.touches[0].clientY - e.touches[1].clientY);
		}
	}

	function onTouchMove(e: TouchEvent) {
		e.preventDefault();
		if (e.touches.length === 2) {
			const dist = Math.hypot(e.touches[0].clientX - e.touches[1].clientX, e.touches[0].clientY - e.touches[1].clientY);
			if (lastPinchDist > 0) zoomBy(lastPinchDist / dist);
			lastPinchDist = dist;
			return;
		}
		if (!dragging || e.touches.length !== 1) return;
		offsetX -= (e.touches[0].clientX - lastX) * scale;
		offsetY -= (e.touches[0].clientY - lastY) * scale;
		lastX = e.touches[0].clientX; lastY = e.touches[0].clientY;
		clampOffset(); draw();
	}

	function confirm() {
		const out = document.createElement('canvas');
		out.width = 1280; out.height = 720;
		out.getContext('2d')!.drawImage(img, offsetX, offsetY, W * scale, H * scale, 0, 0, 1280, 720);
		out.toBlob((blob) => { if (blob) onconfirm(blob, loadedFile); open = false; }, 'image/jpeg', 0.92);
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-3xl">
		<Dialog.Header>
			<Dialog.Title>Crop wallpaper</Dialog.Title>
			<Dialog.Description>Drag to reposition, scroll to zoom.</Dialog.Description>
		</Dialog.Header>

		<div class="flex justify-center my-2">
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<canvas
				bind:this={canvas}
				width={W}
				height={H}
				class="rounded-lg cursor-grab active:cursor-grabbing touch-none w-full"
				onmousedown={onMouseDown}
				onmousemove={onMouseMove}
				onmouseup={onMouseUp}
				onmouseleave={onMouseUp}
				ontouchstart={onTouchStart}
				ontouchmove={onTouchMove}
				ontouchend={() => (dragging = false)}
				onwheel={onWheel}
			></canvas>
		</div>

		<Dialog.Footer>
			<Button variant="outline" onclick={() => (open = false)}>Cancel</Button>
			<Button onclick={confirm} disabled={!imgLoaded}>Use wallpaper</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
