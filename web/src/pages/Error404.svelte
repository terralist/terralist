<script lang="ts">
  import Navbar from '@/components/Navbar.svelte';
  import FullPageError from '@/components/FullPageError.svelte';
</script>

<Navbar />
<main class="mt-36 lg:mt-20 mx-10">
  <FullPageError
    code={404}
    message="This is not the web page you're looking for." />
</main>
