import type { DashboardCommonDictionary } from "../types";

const esDashboardCommonDictionary: DashboardCommonDictionary = {
  sidebar: {
    brandName: "AskAtlas",
    brandTagline: "Tu espacio de estudio",
    groupLabel: "Navegación",
    items: {
      home: "Inicio",
      starred: "Guardados",
      courses: "Cursos",
      studyGuides: "Guías de estudio",
      resources: "Recursos",
      starredAll: "Todos los guardados",
      starredRecent: "Guardados recientes",
      coursesBrowse: "Explorar cursos",
      coursesMine: "Mis cursos",
      coursesSectionMine: "Mis cursos",
      guidesCreate: "Crear nueva guía",
      guidesMine: "Mis guías de estudio",
      guidesRecent: "Vistos recientemente",
      resourcesUpload: "Subir recurso",
      resourcesView: "Ver recursos",
      resourcesRecent: "Vistos recientemente",
      // TODO: Replace these with dynamic samples from the backend
      samples: {
        machineLearningFundamentals: "Fundamentos de Machine Learning",
        binaryTreesCheatSheet: "Guía rápida de árboles binarios",
        neuralNetworksPaper: "Artículo sobre redes neuronales",
        introPsychology: "Introducción a la Psicología",
        dataStructuresAlgorithms: "Estructuras de Datos y Algoritmos",
        modernWebDevelopment: "Desarrollo Web Moderno",
        midtermReviewPsychology: "Repaso de parcial - Psicología",
        algorithmComplexityNotes: "Notas de complejidad algorítmica",
        databasesQuickReference: "Referencia rápida de bases de datos",
        cloudComputingNotes: "Notas de computación en la nube",
      },
    },
  },
  breadcrumb: {
    home: "Inicio",
    browseCourses: "Explorar cursos",
    myCourses: "Mis cursos",
    studyGuides: "Guías de estudio",
    createStudyGuide: "Crear guía de estudio",
    myStudyGuides: "Mis guías de estudio",
    resources: "Recursos",
    uploadResource: "Subir recurso",
    starred: "Guardados",
  },
  userMenu: {
    upgradeToPro: "Mejorar a Pro",
    account: "Cuenta",
    billing: "Facturación",
    notifications: "Notificaciones",
    logout: "Cerrar sesión",
  },
};

export default esDashboardCommonDictionary;
