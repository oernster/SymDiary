// The Guide's words, held apart from the dialog that draws them, so the
// component stays a renderer and the text stays one readable document.
//
// Ported from PigeonPost's guideContent.ts. It does two jobs, in this order. It
// NAMES the furniture, each entry carrying the real picture the band draws, so a
// control can be identified by someone who has just met it. Then it states the
// few rules the window cannot say for itself: what is kept locally, what SymDiary
// will never do and what cannot be undone.

import recordIcon from './assets/icons/record.png'
import historyIcon from './assets/icons/history.png'
import receiptIcon from './assets/icons/receipt.png'
import symptomsIcon from './assets/icons/symptoms.png'
import exportIcon from './assets/icons/export.png'
import importIcon from './assets/icons/import.png'
import lightModeIcon from './assets/icons/light-mode.png'
import guideIcon from './assets/icons/help-info.png'
import crest from './assets/icons/application-icon.png'

/** GuideEntry is one named piece of furniture: its picture, name and what it does. */
export interface GuideEntry {
  icon: string
  name: string
  text: string
}

/** GuideRule is a rule behind the behaviour: a claim, then what it means. */
export interface GuideRule {
  title: string
  text: string
}

/** GuideSection is one block of the document. */
export interface GuideSection {
  heading: string
  intro?: string
  entries?: GuideEntry[]
  rules?: GuideRule[]
  paragraphs?: string[]
}

export const guideSections: GuideSection[] = [
  {
    heading: 'What SymDiary is for',
    paragraphs: [
      'When you notice a symptom, record it. When you see your doctor, take the record with you.',
      'SymDiary keeps what you observed and when you observed it. It does not work out what your symptoms mean: that is for you and your healthcare professional, with the record in front of you both.',
    ],
  },
  {
    heading: 'The bar along the top',
    intro: 'The four screens are on the left; the actions and the Guide are on the right.',
    entries: [
      {
        icon: recordIcon,
        name: 'Record',
        text: 'where SymDiary opens. Type the symptom, add a note or a severity if you want to, then Record now. The time is filled in for you.',
      },
      {
        icon: historyIcon,
        name: 'History',
        text: 'everything recorded, newest first. Narrow it by dates, symptom or severity; correct or delete an entry here.',
      },
      {
        icon: receiptIcon,
        name: 'Receipt',
        text: 'the record for a range of dates, to take to an appointment. Mark the date of your last appointment and the range starts there next time; choose Custom dates to pick your own. When the range covers that date, the sheet names it under the dates; untick Show on the sheet to leave it off. Save PDF writes it to a file, numbered page by page, which you can print or email.',
      },
      {
        icon: symptomsIcon,
        name: 'Symptoms',
        text: 'the symptoms you have used so far. Rename one here and every entry using it follows.',
      },
      {
        icon: importIcon,
        name: 'Import',
        text: 'reads an exported file back in. Anything already in your record is skipped, so importing the same file twice changes nothing.',
      },
      {
        icon: exportIcon,
        name: 'Export',
        text: 'writes your whole record to a file you choose, in the Downloads folder unless you pick another. The file is plain JSON: it is yours.',
      },
      {
        icon: lightModeIcon,
        name: 'Light mode',
        text: 'moves the window between dark and light. The picture is the one you would move to, so the sun means light is a press away; SymDiary opens in whichever you left it in.',
      },
      {
        icon: guideIcon,
        name: 'Guide',
        text: 'this document.',
      },
      {
        icon: crest,
        name: 'About',
        text: 'the version, who wrote it and the open-source works it is built with.',
      },
    ],
  },
  {
    heading: 'Recording quickly',
    rules: [
      {
        title: 'A few keys is the whole thing.',
        text: 'Type the first letters of a symptom you have used before, press the down arrow to reach it, Enter to take it, then Enter again to record.',
      },
      {
        title: 'A symptom is yours to name.',
        text: 'Type anything. The first time you use a name it is kept for next time, exactly as you typed it. SymDiary never rewords it into medical terms.',
      },
      {
        title: 'Severity and notes are optional.',
        text: 'Leave either out. Nothing asks you to put a number on something you do not think of that way.',
      },
      {
        title: 'Something you remember later still belongs in the record.',
        text: 'Set a different time on the recording screen and enter it then. SymDiary keeps both when it happened and when you wrote it down.',
      },
    ],
  },
  {
    heading: 'What SymDiary will not do',
    rules: [
      {
        title: 'It does not diagnose.',
        text: 'No suggested conditions, no likely causes and no advice about medication or treatment.',
      },
      {
        title: 'It does not interpret.',
        text: 'No trends, scores, ratings or claims that one thing led to another. The receipt counts your entries and shows them; that is all the arithmetic there is.',
      },
      {
        title: 'It does not nag.',
        text: 'No reminders and no prompts to record.',
      },
    ],
  },
  {
    heading: 'Your record',
    rules: [
      {
        title: 'It stays on this computer.',
        text: "There is no account, no cloud service and no advertising. SymDiary itself opens no network connection; exporting and saving a PDF happen only when you ask. The one exception never touches your record: on a Windows computer missing Microsoft's WebView2, which the window is drawn with, SymDiary asks before fetching Microsoft's installer for it.",
      },
      {
        title: 'It is not encrypted.',
        text: 'The file is protected by the account you sign in to, as your documents are. Anyone who can sign in as you can read it.',
      },
      {
        title: 'Deleting cannot be undone.',
        text: 'SymDiary asks first and names what will go. Once it is gone, only an export made earlier can bring it back.',
      },
    ],
  },
]
