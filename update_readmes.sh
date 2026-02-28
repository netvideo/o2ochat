#!/bin/bash

# Update all README files with legal warning

# Spanish
sed -i '/^---$/,/<p align="center">/c\
---\
\
### ⚠️ Advertencia Legal\
\
**Aviso Importante: Este proyecto es solo para fines educativos**\
\
- 📚 **Propósito Educativo** - Demuestra comunicación P2P y cifrado\
- ⚖️ **Cumplimiento Legal** - Debe cumplir con leyes locales\
- 🚫 **Prohibido Uso Ilegal** - No para actividades ilegales\
- 📝 **Responsabilidad** - Usuario tiene responsabilidad legal\
- 🔒 **Neutralidad Tecnológica** - La tecnología es neutral\
\
**Al usar este proyecto, aceptas:**\
1. Usar solo para propósitos legales\
2. No realizar actividades ilegales\
3. Aceptar riesgos técnicos\
4. Cumplir [Términos](TERMS_OF_SERVICE.md) y [Privacidad](PRIVACY.md)\
\
Ver: [Aviso de Seguridad](SECURITY_NOTICE.md)\
\
---\
\
<p align="center">\
  <b>P2P Puro | Cifrado Extremo a Extremo | Comunicación Libre</b>\
</p>' README_ES.md

echo "Updated README_ES.md"
