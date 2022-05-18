const fs = require('fs');

const holder_color = '#EEEEEE';
const colors = [
  '#7CB342',
  '#AED581',
  '#DCE775',
  '#CDDC39',
  '#FFA726',
  '#FB8C00',
  '#EF6C00',
  '#E65100',
  '#E53935',
  '#B71C1C',
  '#D50000',
];

for (let risk = 0; risk <= 10; risk++) {
  const riskLevel = Math.ceil(risk / 2);
  const color = colors[risk];
  let output = `
        <svg xmlns="http://www.w3.org/2000/svg" width="52" height="10">
          <circle cx="5" cy="5" r="5"  fill="${
            riskLevel >= 1 ? color : holder_color
          }"/>
          <circle cx="15.5" cy="5" r="5" fill="${
            riskLevel >= 2 ? color : holder_color
          }"/>
          <circle cx="26" cy="5" r="5" fill="${
            riskLevel >= 3 ? color : holder_color
          }"/>
          <circle cx="36.5" cy="5" r="5" fill="${
            riskLevel >= 4 ? color : holder_color
          }"/>
          <circle cx="47" cy="5" r="5" fill="${
            riskLevel >= 5 ? color : holder_color
          }"/>
        </svg>
    `;
  output = output.trim();
  output = output.replace(/ {8}/g, '');
  fs.writeFileSync(`./src/repositories/svg/scores/${risk}.svg`, output);

  // minify
  output = output.replace(/\n/g, '');
  output = output.replace(/>\s*</g, '><');
  fs.writeFileSync(`./public/svg/scores/${risk}.svg`, output);
}
