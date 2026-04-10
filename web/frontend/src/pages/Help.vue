<script setup>
import { ref } from 'vue'

const openGuide = ref(null)
const openFaq = ref(null)

function toggleGuide(i) { openGuide.value = openGuide.value === i ? null : i }
function toggleFaq(i) { openFaq.value = openFaq.value === i ? null : i }

const guides = [
  { title: 'Inscrire un patient', content: "Allez dans Patients puis cliquez sur \"+ Nouveau patient\". Remplissez le formulaire avec le nom, prenom, sexe et les informations de contact. Le code patient sera genere automatiquement." },
  { title: 'Programmer un rendez-vous', content: "Allez dans Calendrier puis cliquez \"+ RDV\", ou depuis le menu principal. Recherchez le patient, choisissez le type de consultation, la date et un creneau disponible." },
  { title: 'Consulter le calendrier', content: "Le calendrier propose deux vues : Semaine (grille horaire de 8h a 16h) et Mois (vue globale avec compteurs par jour). Utilisez les fleches pour naviguer." },
  { title: 'Gerer les rappels', content: "La page Rappels affiche les statistiques d'envoi, la file d'attente des messages et les modeles. Vous pouvez modifier les modeles de SMS et lancer un envoi grouppe." },
  { title: 'Dossier patient', content: "Cliquez sur un patient dans la liste pour ouvrir son dossier. Vous y trouverez ses informations, son score de risque, et les actions disponibles (programmer un RDV, sortie du programme)." },
  { title: 'Exporter des rapports', content: "Depuis la page Parametres, telechargez le rapport mensuel en Excel ou PDF, ou exportez la liste des patients actifs ou perdus de vue." },
  { title: 'Gerer les utilisateurs', content: "Les administrateurs peuvent creer de nouveaux comptes depuis Parametres > Gestion des utilisateurs. Chaque utilisateur a un role (medecin, infirmier, etc.) qui determine ses acces." },
  { title: 'Modifier son profil', content: "Depuis Mon profil (accessible via le menu), modifiez votre nom, email et telephone. Vous pouvez aussi changer votre mot de passe (redirection vers la page de connexion apres changement)." },
]

const faqs = [
  { q: 'Comment retrouver un patient ?', a: "Utilisez la barre de recherche en haut de la page ou allez dans Patients. Vous pouvez chercher par nom, prenom ou code patient." },
  { q: "Que signifie le score de risque ?", a: "Le score de risque (0-10) evalue la probabilite de perte de suivi du patient. Faible (0-3): suivi regulier. Moyen (4-6): a surveiller. Eleve (7-10): intervention urgente necessaire." },
  { q: "Comment fonctionne l'envoi de rappels ?", a: "Les rappels sont programmes automatiquement avant chaque rendez-vous. Ils sont envoyes par SMS, WhatsApp ou appel selon la preference du patient. Vous pouvez aussi lancer un envoi manuel depuis la page Rappels." },
  { q: 'Puis-je reinscire un patient sorti ?', a: "Non, une fois sorti du programme, le dossier est cloture. Vous devez creer une nouvelle inscription si le patient revient." },
  { q: "Comment changer mon mot de passe ?", a: "Allez dans Mon profil, section \"Changer le mot de passe\". Le nouveau mot de passe doit contenir au moins 8 caracteres et un chiffre. Vous serez redirige vers la page de connexion." },
  { q: 'Quels navigateurs sont supportes ?', a: "MaSante fonctionne sur les versions recentes de Chrome, Firefox, Safari et Edge. Un acces Internet est necessaire." },
]
</script>

<template>
  <div class="grid-2" style="align-items:start">
    <!-- Guides -->
    <div class="card">
      <div class="card-head"><h3>Guide d'utilisation</h3></div>
      <div class="card-body">
        <div v-for="(g, i) in guides" :key="i" style="border-bottom:1px solid var(--gray-50)">
          <div style="display:flex;align-items:center;justify-content:space-between;padding:12px 0;cursor:pointer;gap:8px" @click="toggleGuide(i)">
            <div style="font-weight:600;font-size:.88rem">{{ g.title }}</div>
            <span style="font-size:.8rem;color:var(--gray-400);flex-shrink:0;transition:transform .2s" :style="openGuide === i ? 'transform:rotate(180deg)' : ''">&#9660;</span>
          </div>
          <div v-if="openGuide === i" style="padding:0 0 14px;font-size:.84rem;color:var(--gray-500);line-height:1.5">{{ g.content }}</div>
        </div>
      </div>
    </div>

    <!-- FAQ -->
    <div class="card">
      <div class="card-head"><h3>Questions frequentes</h3></div>
      <div class="card-body">
        <div v-for="(f, i) in faqs" :key="i" style="border-bottom:1px solid var(--gray-50)">
          <div style="display:flex;align-items:center;justify-content:space-between;padding:12px 0;cursor:pointer;gap:8px" @click="toggleFaq(i)">
            <div style="font-weight:600;font-size:.88rem">{{ f.q }}</div>
            <span style="font-size:.8rem;color:var(--gray-400);flex-shrink:0;transition:transform .2s" :style="openFaq === i ? 'transform:rotate(180deg)' : ''">&#9660;</span>
          </div>
          <div v-if="openFaq === i" style="padding:0 0 14px;font-size:.84rem;color:var(--gray-500);line-height:1.5">{{ f.a }}</div>
        </div>
      </div>
    </div>
  </div>
</template>
