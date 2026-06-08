import { createApp } from "vue";
import PrimeVue from "primevue/config";
import App from "./App.vue";
import { router } from "./router";
import { primeVueConfig } from "./ui/primevue/config";
import "./styles/global.css";

createApp(App).use(router).use(PrimeVue, primeVueConfig).mount("#app");
