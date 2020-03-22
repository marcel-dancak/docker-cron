import Vue from 'vue'
import { VIcon } from 'vuetify/lib'

import DateField from '@/components/DateField'
import TimeField from '@/components/TimeField'
import Icon from '@/components/Icon'

// import all icons
const svgIcons = require.context('../icons', false, /.*\.svg$/)
svgIcons.keys().map(svgIcons)

Vue.component('icon', Icon)
Vue.component('v-icon', VIcon)
Vue.component('date-field', DateField)
Vue.component('time-field', TimeField)
