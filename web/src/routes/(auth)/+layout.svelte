<script lang="ts">
	let { children } = $props();
</script>

<div class="min-h-screen flex items-center justify-center bg-background px-4 py-12">
	<div class="w-full max-w-sm">
		<div class="text-center mb-8">
			<span class="text-4xl">🏠</span>
			<h1 class="text-2xl font-bold mt-2 text-foreground">Household</h1>
			<p class="text-sm text-muted-foreground mt-1">Your household, organised.</p>
		</div>
		<div class="bg-card border border-border rounded-xl p-6 shadow-sm">
			{@render children()}
		</div>
	</div>
</div>
