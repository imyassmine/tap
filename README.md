Protocol Implementation
Ajouts par rapport au RFC : ERR 400 BAD_REQUEST, ERR 202 NOT_AUTHENTICATED, les lignes vides ignorées, les règles de validité des pseudos, et le CONNECT répété qui renvoie 400.
Un item ou un NPC ne peut être que dans une seule salle : le sujet demande des items comme « instances uniques ». Un item listé dans deux salles serait dupliqué dès le démarrage.
Les directions sont en minuscules, sans espace. Le RFC ne définit pas de liste de directions (MOVE <direction>, avec 301 NO_EXIT pour une direction invalide), donc ce sont les clés du fichier. Le serveur mettra l’argument du client en minuscules avant de chercher la sortie, pour que MOVE North fonctionne.
Un item non placé dans une salle est accepté, parce qu’il peut être une récompense de quête.