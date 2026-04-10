<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { patients } from '../api'
import { useToast } from '../composables/useToast'

const router = useRouter()
const toast = useToast()
const form = ref({ last_name:'',first_name:'',date_of_birth:'',sex:'',phone:'',phone_secondary:'',district:'',address:'',language:'fr',reminder_channel:'sms',contact_name:'',contact_phone:'',contact_relation:'',referred_by:'' })
const error = ref('')
const saving = ref(false)

async function submit() {
  error.value = ''
  if (!form.value.last_name) { error.value = 'Le nom est requis'; return }
  if (!form.value.first_name) { error.value = 'Le prenom est requis'; return }
  if (!form.value.sex) { error.value = 'Le sexe est requis'; return }
  saving.value = true
  const res = await patients.create(form.value)
  saving.value = false
  if (!res.ok) { error.value = res.error; return }
  toast.success('Patient inscrit — Code: ' + (res.data.Code || ''))
  router.push('/patients')
}
</script>

<template>
  <div class="apt-grid">
    <div>
      <div class="card">
        <div class="card-head"><h3>Inscription</h3></div>
        <div class="card-body">
          <div style="display:grid;grid-template-columns:1fr 1fr;gap:14px">
            <div class="form-group"><label>Nom *</label><input class="form-input" v-model="form.last_name" placeholder="Nom de famille"></div>
            <div class="form-group"><label>Prenom *</label><input class="form-input" v-model="form.first_name" placeholder="Prenom"></div>
            <div class="form-group"><label>Date de naissance</label><input class="form-input" type="date" v-model="form.date_of_birth"></div>
            <div class="form-group"><label>Sexe *</label><select class="form-input" v-model="form.sex"><option value="">Selectionner</option><option value="M">Masculin</option><option value="F">Feminin</option></select></div>
            <div class="form-group"><label>Telephone</label><input class="form-input" v-model="form.phone" placeholder="+237 6XX XXX XXX"></div>
            <div class="form-group"><label>Tel secondaire</label><input class="form-input" v-model="form.phone_secondary" placeholder="Optionnel"></div>
            <div class="form-group"><label>Quartier</label><input class="form-input" v-model="form.district" placeholder="Ex: Akwa"></div>
            <div class="form-group"><label>Adresse</label><input class="form-input" v-model="form.address" placeholder="Repere"></div>
          </div>
          <div class="form-group" style="margin-top:14px"><label>Langue</label>
            <div class="remind-opts">
              <div v-for="[l,c] in [['Francais','fr'],['Anglais','en'],['Duala','duala'],['Ewondo','ewondo'],['Bamileke','bamileke']]" :key="c" class="r-opt" :class="{on:form.language===c}" @click="form.language=c">{{l}}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div>
      <div class="card" style="margin-bottom:16px">
        <div class="card-head"><h3>Contact et rappels</h3></div>
        <div class="card-body">
          <div class="form-group"><label>Canal de rappel</label>
            <div class="remind-opts">
              <div v-for="[l,c] in [['SMS','sms'],['WhatsApp','whatsapp'],['Appel','voice'],['Aucun','none']]" :key="c" class="r-opt" :class="{on:form.reminder_channel===c}" @click="form.reminder_channel=c">{{l}}</div>
            </div>
          </div>
          <div class="form-group"><label>Personne de confiance</label><input class="form-input" v-model="form.contact_name" placeholder="Nom"></div>
          <div class="form-group"><label>Tel contact</label><input class="form-input" v-model="form.contact_phone" placeholder="+237"></div>
          <div class="form-group"><label>Lien</label><select class="form-input" v-model="form.contact_relation"><option v-for="r in ['','Conjoint(e)','Parent','Frere/Soeur','Ami(e)','Autre']" :key="r" :value="r">{{r||'Selectionner'}}</option></select></div>
          <div class="form-group"><label>Refere par</label><select class="form-input" v-model="form.referred_by"><option v-for="r in ['','Centre de depistage','Transfert','Auto-presentation','Agent communautaire','Autre']" :key="r" :value="r">{{r||'Selectionner'}}</option></select></div>
        </div>
      </div>
      <div v-if="error" style="padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin-bottom:12px">{{error}}</div>
      <button class="btn btn-primary" @click="submit" :disabled="saving">{{saving?'Inscription...':'Inscrire le patient'}}</button>
    </div>
  </div>
</template>
