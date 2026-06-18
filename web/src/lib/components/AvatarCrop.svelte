<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';

	let {
		open = $bindable(false),
		onconfirm
	}: {
		open: boolean;
		onconfirm: (blob: Blob) => void;
	} = $props();

	let canvas: HTMLCanvasElement;
	let img = new Image();
	let imgLoaded = $state(false);

	// Crop state: offset is the top-left corner of the crop square in image coords
	let scale = $state(1);       // zoom: crop square covers (canvas.width / scale) px of image
	let offsetX = $state(0);     // pan offset in image pixels
	let offsetY = $state(0);

	let dragging = false;
	let lastX = 0;
	let lastY = 0;

	export function loadFile(file: File) {
		const reader = new FileReader();
		reader.onload = (e) => {
			img.onload = () => {
				imgLoaded = true;
				// Fit the image so the shorter dimension fills the canvas
				const canvasSize = 320;
				scale = Math.max(img.naturalWidth, img.naturalHeight) / canvasSize;
				offsetX = (img.naturalWidth - canvasSize * scale) / 2;
				offsetY = (img.naturalHeight - canvasSize * scale) / 2;
				draw();
			};
			img.src = e.target!.result as string;
		};
		reader.readAsDataURL(file);
	}

	function draw() {
		if (!canvas || !imgLoaded) return;
		const ctx = canvas.getContext('2d')!;
		const s = canvas.width; // square canvas

		ctx.clearRect(0, 0, s, s);
		ctx.drawImage(img,
			offsetX, offsetY, s * scale, s * scale,
			0, 0, s, s
		);

		// Darken outside circle
		ctx.save();
		ctx.fillStyle = 'rgba(0,0,0,0.45)';
		ctx.fillRect(0, 0, s, s);
		ctx.globalCompositeOperation = 'destination-out';
		ctx.beginPath();
		ctx.arc(s / 2, s / 2, s / 2 - 4, 0, Math.PI * 2);
		ctx.fill();
		ctx.restore();

		// Circle border
		ctx.save();
		ctx.strokeStyle = 'rgba(255,255,255,0.8)';
		ctx.lineWidth = 2;
		ctx.beginPath();
		ctx.arc(s / 2, s / 2, s / 2 - 4, 0, Math.PI * 2);
		ctx.stroke();
		ctx.restore();
	}

	function clampOffset() {
		const s = canvas?.width ?? 320;
		const maxX = img.naturalWidth - s * scale;
		const maxY = img.naturalHeight - s * scale;
		offsetX = Math.max(0, Math.min(offsetX, maxX));
		offsetY = Math.max(0, Math.min(offsetY, maxY));
	}

	function onMouseDown(e: MouseEvent) {
		dragging = true;
		lastX = e.clientX;
		lastY = e.clientY;
	}

	function onMouseMove(e: MouseEvent) {
		if (!dragging) return;
		offsetX -= (e.clientX - lastX) * scale;
		offsetY -= (e.clientY - lastY) * scale;
		lastX = e.clientX;
		lastY = e.clientY;
		clampOffset();
		draw();
	}

	function onMouseUp() { dragging = false; }

	function onTouchStart(e: TouchEvent) {
		if (e.touches.length === 1) {
			dragging = true;
			lastX = e.touches[0].clientX;
			lastY = e.touches[0].clientY;
		}
	}

	function onTouchMove(e: TouchEvent) {
		e.preventDefault();
		if (!dragging || e.touches.length !== 1) return;
		offsetX -= (e.touches[0].clientX - lastX) * scale;
		offsetY -= (e.touches[0].clientY - lastY) * scale;
		lastX = e.touches[0].clientX;
		lastY = e.touches[0].clientY;
		clampOffset();
		draw();
	}

	function onWheel(e: WheelEvent) {
		e.preventDefault();
		const s = canvas.width;
		const delta = e.deltaY > 0 ? 1.1 : 0.9;
		const newScale = Math.max(0.5, Math.min(scale * delta, Math.max(img.naturalWidth, img.naturalHeight) / s));
		// Zoom around canvas center
		const cx = s / 2;
		const cy = s / 2;
		offsetX = offsetX + cx * (scale - newScale);
		offsetY = offsetY + cy * (scale - newScale);
		scale = newScale;
		clampOffset();
		draw();
	}

	function confirm() {
		// Output a 256×256 canvas (circle already applied by the overlay)
		const out = document.createElement('canvas');
		out.width = 256;
		out.height = 256;
		const ctx = out.getContext('2d')!;
		// Clip to circle
		ctx.beginPath();
		ctx.arc(128, 128, 128, 0, Math.PI * 2);
		ctx.clip();
		ctx.drawImage(img, offsetX, offsetY, canvas.width * scale, canvas.width * scale, 0, 0, 256, 256);
		out.toBlob((blob) => {
			if (blob) onconfirm(blob);
			open = false;
		}, 'image/jpeg', 0.9);
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-sm">
		<Dialog.Header>
			<Dialog.Title>Crop photo</Dialog.Title>
			<Dialog.Description>Drag to reposition, scroll to zoom.</Dialog.Description>
		</Dialog.Header>

		<div class="flex justify-center my-2">
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<canvas
				bind:this={canvas}
				width="320"
				height="320"
				class="rounded-full cursor-grab active:cursor-grabbing touch-none"
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
			<Button onclick={confirm} disabled={!imgLoaded}>Use photo</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
