<template> 
  <div class="shadow-sm bg-white rounded-md flex flex-col justify-center items-center py-6">
    <img :src="this.imageSrc" class="w-24" />
    <h3 class="font-display text-lg text-text font-medium pt-4">{{ title }}</h3>
    <h6 class="font-display text-md text-text">{{ subtitle }}</h6>
  </div>
</template>

<script>
export default {
  name: 'EmptyFiller',
  props: {
    imageSrc: String,
    title: String,
    subtitle: String
  },
}
</script>

<style>

</style>
